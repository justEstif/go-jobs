package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// LLMClientFactory constructs an LLMClient from user settings.
// Injected by the composition root so the service layer never imports adapters.
type LLMClientFactory func(provider domain.LLMProvider, apiKey, model, baseURL string) (ports.LLMClient, error)

// KeyDecryptor decrypts an encrypted API key. Injected by the composition root.
type KeyDecryptor func(ciphertext string) (string, error)

// coachService implements ports.JobCoachService.
type coachService struct {
	users      ports.UserRepository
	jobs       ports.JobRepository
	companies  ports.CompanyRepository
	cache      ports.CoachCacheRepository
	newClient  LLMClientFactory
	decryptKey KeyDecryptor
}

// NewJobCoachService constructs a JobCoachService.
func NewJobCoachService(
	users ports.UserRepository,
	jobs ports.JobRepository,
	companies ports.CompanyRepository,
	cache ports.CoachCacheRepository,
	newClient LLMClientFactory,
	decryptKey KeyDecryptor,
) ports.JobCoachService {
	return &coachService{
		users:      users,
		jobs:       jobs,
		companies:  companies,
		cache:      cache,
		newClient:  newClient,
		decryptKey: decryptKey,
	}
}

// AnalyzeJob compares the user's resume against a job posting and returns
// structured analysis with ATS optimization and an optimized resume.
// Returns a cached result if available unless refresh is true.
func (s *coachService) AnalyzeJob(ctx context.Context, userID domain.UserID, jobID domain.JobID, refresh bool) (string, error) {
	// Check cache first (unless refresh requested).
	if !refresh {
		cached, err := s.cache.Get(ctx, userID, jobID, domain.CoachKindAnalyze)
		if err == nil {
			return cached.Result, nil
		}
		// Cache miss — continue to LLM call.
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("coach: get user: %w", err)
	}
	if user.Resume == "" {
		return "", fmt.Errorf("no resume configured — add your resume in Settings")
	}
	if user.LLMProvider == "" {
		return "", fmt.Errorf("no LLM provider configured — set up a provider in Settings")
	}

	job, err := s.jobs.GetByID(ctx, jobID)
	if err != nil {
		return "", fmt.Errorf("coach: get job: %w", err)
	}

	company, err := s.companies.GetByID(ctx, job.CompanyID)
	if err != nil {
		return "", fmt.Errorf("coach: get company: %w", err)
	}

	client, err := s.buildClient(user)
	if err != nil {
		return "", err
	}

	systemPrompt := buildAnalyzeSystemPrompt()
	userPrompt := buildAnalyzeUserPrompt(user.Resume, job, company)

	result, err := client.Complete(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("coach: llm complete: %w", err)
	}

	// Cache the result.
	modelUsed := fmt.Sprintf("%s/%s", user.LLMProvider, user.LLMModel)
	if err := s.cache.Upsert(ctx, domain.CoachCache{
		UserID:    userID,
		JobID:     jobID,
		Kind:      domain.CoachKindAnalyze,
		Result:    result,
		ModelUsed: modelUsed,
	}); err != nil {
		log.Printf("coach: failed to cache analysis: %v", err)
		// Non-fatal — return the result anyway.
	}

	return result, nil
}

// GenerateCaseStudy expands a project description into a structured case study.
func (s *coachService) GenerateCaseStudy(ctx context.Context, userID domain.UserID, projectDescription string) (string, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("coach: get user: %w", err)
	}
	if user.LLMProvider == "" {
		return "", fmt.Errorf("no LLM provider configured — set up a provider in Settings")
	}
	if projectDescription == "" {
		return "", fmt.Errorf("project description is required")
	}

	client, err := s.buildClient(user)
	if err != nil {
		return "", err
	}

	systemPrompt := buildCaseStudySystemPrompt()
	userPrompt := buildCaseStudyUserPrompt(user.Resume, projectDescription)

	result, err := client.Complete(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("coach: llm complete: %w", err)
	}

	return result, nil
}

// BuildAnalyzePrompt returns the raw prompts for job analysis without calling the LLM.
// Only requires a resume — no LLM provider needed.
func (s *coachService) BuildAnalyzePrompt(ctx context.Context, userID domain.UserID, jobID domain.JobID) (string, string, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("coach: get user: %w", err)
	}
	if user.Resume == "" {
		return "", "", fmt.Errorf("no resume configured — add your resume in Settings")
	}

	job, err := s.jobs.GetByID(ctx, jobID)
	if err != nil {
		return "", "", fmt.Errorf("coach: get job: %w", err)
	}

	company, err := s.companies.GetByID(ctx, job.CompanyID)
	if err != nil {
		return "", "", fmt.Errorf("coach: get company: %w", err)
	}

	return buildAnalyzeSystemPrompt(), buildAnalyzeUserPrompt(user.Resume, job, company), nil
}

// BuildCaseStudyPrompt returns the raw prompts for case study generation.
func (s *coachService) BuildCaseStudyPrompt(ctx context.Context, userID domain.UserID, projectDescription string) (string, string, error) {
	if projectDescription == "" {
		return "", "", fmt.Errorf("project description is required")
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("coach: get user: %w", err)
	}

	return buildCaseStudySystemPrompt(), buildCaseStudyUserPrompt(user.Resume, projectDescription), nil
}

// buildClient constructs an LLMClient from the user's settings, decrypting
// the API key if needed.
func (s *coachService) buildClient(user domain.User) (ports.LLMClient, error) {
	apiKey := ""
	if user.LLMProvider != domain.LLMOllama {
		var err error
		apiKey, err = s.decryptKey(user.LLMAPIKey)
		if err != nil {
			return nil, fmt.Errorf("coach: decrypt api key: %w", err)
		}
		if apiKey == "" {
			return nil, fmt.Errorf("no API key configured for %s — add your key in Settings", user.LLMProvider)
		}
	}

	client, err := s.newClient(user.LLMProvider, apiKey, user.LLMModel, user.LLMBaseURL)
	if err != nil {
		return nil, fmt.Errorf("coach: create llm client: %w", err)
	}
	return client, nil
}

// buildAnalyzeSystemPrompt returns the system prompt for job analysis.
func buildAnalyzeSystemPrompt() string {
	return `You are a senior career coach and ATS (Applicant Tracking System) optimization expert. You analyze job postings against a candidate's resume and provide actionable, specific feedback.

Your analysis MUST follow this exact structure with these section headers:

## ATS Analysis
Analyze keyword gaps between the resume and job description. Every point must reference a specific term from the JD that is missing or could be better represented in the resume. Include formatting advice specific to the ATS platform if provided.

## Fit Analysis
Assess the candidate's fit for this role. List specific strengths (skills/experience they have that match requirements) and specific gaps (requirements they don't clearly meet). Classify each gap as:
- **Critical gap**: Likely disqualifying — missing a hard requirement
- **Major gap**: Significant but addressable in cover letter or interview
- **Minor gap**: Easy to learn or can frame existing experience

## Optimized Resume
Generate a complete, tailored version of the candidate's resume optimized for this specific job. The resume must:
- Use the job description's exact terminology where the candidate has matching experience
- Prioritize and expand experience most relevant to this role
- De-emphasize or condense irrelevant experience
- Include quantified achievements where possible
- Be formatted with clear section headers (Summary, Experience, Skills, Education)
- Only include skills and experience that are actually present in the original resume — never fabricate

CRITICAL RULES:
- Every bullet point in ATS Analysis and Fit Analysis must reference specific text from either the resume or job description.
- Never give generic advice like "tailor your resume" or "highlight relevant experience."
- Never fabricate skills or experience the candidate doesn't have.
- Be honest about gaps — constructive honesty is more valuable than false encouragement.
- The Optimized Resume must be a complete, ready-to-submit document the candidate can copy-paste.`
}

// buildAnalyzeUserPrompt constructs the user prompt with resume, job, and company context.
func buildAnalyzeUserPrompt(resume string, job domain.Job, company domain.Company) string {
	var b strings.Builder

	b.WriteString("## Candidate Resume\n\n")
	b.WriteString(resume)
	b.WriteString("\n\n---\n\n")

	b.WriteString("## Job Posting\n\n")
	b.WriteString(fmt.Sprintf("**Title:** %s\n", job.Title))
	b.WriteString(fmt.Sprintf("**Company:** %s\n", company.Name))
	b.WriteString(fmt.Sprintf("**Location:** %s\n", job.Location))
	if job.URL != "" {
		b.WriteString(fmt.Sprintf("**URL:** %s\n", job.URL))
	}
	b.WriteString("\n### Description\n\n")
	b.WriteString(job.Description)

	// Include enriched tags if available for additional context.
	if job.Tags != nil {
		b.WriteString("\n\n### Enriched Metadata (from our system)\n")
		if job.Tags.RoleType != "" {
			b.WriteString(fmt.Sprintf("- Role Type: %s\n", job.Tags.RoleType))
		}
		if job.Tags.Seniority != "" {
			b.WriteString(fmt.Sprintf("- Seniority: %s\n", job.Tags.Seniority))
		}
		if job.Tags.RemotePolicy != "" {
			b.WriteString(fmt.Sprintf("- Remote Policy: %s\n", job.Tags.RemotePolicy))
		}
		if len(job.Tags.TechStack) > 0 {
			b.WriteString(fmt.Sprintf("- Tech Stack: %s\n", strings.Join(job.Tags.TechStack, ", ")))
		}
	}

	b.WriteString(fmt.Sprintf("\n\n### ATS Platform: %s\n", strings.ToUpper(string(company.ATSType))))
	b.WriteString("Provide ATS-specific formatting and keyword advice for this platform.\n")

	return b.String()
}

// buildCaseStudySystemPrompt returns the system prompt for case study generation.
func buildCaseStudySystemPrompt() string {
	return `You are an expert portfolio writer who transforms project descriptions and resume bullets into detailed, compelling case studies.

Generate a structured case study following this exact format:

## Overview
A 2-3 sentence summary of the project including your role, timeline, and one-line impact statement.

## Problem
What needed to be solved? Include business context, user pain points, key metrics or goals, and constraints. Make this specific and grounded — not vague.

## Process
How did you approach it? Include:
- Research conducted
- Options considered
- Key decisions made and why
- Stakeholders involved

## Solution
What did you build/create? Describe the key features or changes. Be specific about technical choices and implementation details. Include descriptions of where visual artifacts (diagrams, screenshots) would go.

## Results
Quantified impact with specific metrics. Use a before/after table format where possible. Include both primary metrics and secondary effects.

## Learnings
- What worked well
- What you'd do differently
- Skills developed

CRITICAL RULES:
- If the project description is brief, expand thoughtfully but don't invent facts. Ask the reader to fill in specifics with [PLACEHOLDER: describe X] markers.
- Use concrete numbers and metrics where provided. If not provided, use [X%] placeholders.
- The case study should be 3-5 minutes to read.
- Write in first person.
- Focus on YOUR specific contributions, not just the team's work.`
}

// buildCaseStudyUserPrompt constructs the user prompt for case study generation.
func buildCaseStudyUserPrompt(resume, projectDescription string) string {
	var b strings.Builder

	b.WriteString("## Project to Expand into a Case Study\n\n")
	b.WriteString(projectDescription)

	if resume != "" {
		b.WriteString("\n\n---\n\n")
		b.WriteString("## Full Resume (for context on my background)\n\n")
		b.WriteString(resume)
	}

	b.WriteString("\n\nGenerate a detailed portfolio case study from this project description. Use my resume for context about my skills and background, but focus the case study on this specific project.")

	return b.String()
}
