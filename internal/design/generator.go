package design

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mojomast/geoffrussy/internal/provider"
	"github.com/mojomast/geoffrussy/internal/state"
)

// Generator generates system architecture from interview data
type Generator struct {
	provider provider.Provider
	model    string
}

// NewGenerator creates a new design generator
func NewGenerator(provider provider.Provider, model string) *Generator {
	return &Generator{
		provider: provider,
		model:    model,
	}
}

// Architecture represents the system architecture
type Architecture struct {
	ProjectID         string
	SystemOverview    string
	Components        []Component
	DataFlows         []DataFlow
	TechRationale     map[string]string
	ScalingStrategy   ScalingPlan
	APIContract       APISpec
	DatabaseSchema    Schema
	SecurityApproach  SecurityPlan
	Observability     ObservabilityPlan
	Deployment        DeploymentPlan
	Risks             []Risk
	Assumptions       []string
	Unknowns          []string
	CreatedAt         time.Time
}

// Component represents a system component
type Component struct {
	Name         string
	Type         ComponentType
	Purpose      string
	Technologies []string
	Dependencies []string
}

// ComponentType represents the type of component
type ComponentType string

const (
	ComponentFrontend   ComponentType = "frontend"
	ComponentBackend    ComponentType = "backend"
	ComponentDatabase   ComponentType = "database"
	ComponentCache      ComponentType = "cache"
	ComponentQueue      ComponentType = "queue"
	ComponentMonitoring ComponentType = "monitoring"
)

// DataFlow represents a data flow through the system
type DataFlow struct {
	Name        string
	Description string
	Steps       []FlowStep
	Diagram     string
}

// FlowStep represents a step in a data flow
type FlowStep struct {
	Order       int
	Component   string
	Action      string
	Description string
}

// ScalingPlan describes how the system scales
type ScalingPlan struct {
	HorizontalScaling string
	VerticalScaling   string
	Caching           string
	LoadBalancing     string
	DatabaseScaling   string
}

// APISpec describes the API contract
type APISpec struct {
	RESTEndpoints []Endpoint
	WebSockets    []WebSocketEvent
	Authentication string
}

// Endpoint represents a REST endpoint
type Endpoint struct {
	Method      string
	Path        string
	Description string
	Request     string
	Response    string
}

// WebSocketEvent represents a WebSocket event
type WebSocketEvent struct {
	Name        string
	Direction   string
	Description string
	Payload     string
}

// Schema represents the database schema
type Schema struct {
	Tables       []Table
	Relationships []Relationship
}

// Table represents a database table
type Table struct {
	Name        string
	Description string
	Columns     []Column
}

// Column represents a table column
type Column struct {
	Name        string
	Type        string
	Constraints string
}

// Relationship represents a table relationship
type Relationship struct {
	From string
	To   string
	Type string
}

// SecurityPlan describes the security approach
type SecurityPlan struct {
	Authentication string
	Authorization  string
	Encryption     string
	Audit          string
}

// ObservabilityPlan describes the observability strategy
type ObservabilityPlan struct {
	Logging string
	Metrics string
	Tracing string
}

// DeploymentPlan describes the deployment architecture
type DeploymentPlan struct {
	Development string
	Staging     string
	Production  string
}

// Risk represents a potential risk
type Risk struct {
	Name        string
	Probability RiskLevel
	Impact      RiskLevel
	Mitigation  string
}

// RiskLevel represents the level of a risk
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// GenerateArchitecture generates a complete system architecture from interview data
func (g *Generator) GenerateArchitecture(interviewData *state.InterviewData) (*Architecture, error) {
	if g.provider == nil {
		return nil, fmt.Errorf("provider is required for architecture generation")
	}

	// Create the architecture prompt
	prompt := g.buildArchitecturePrompt(interviewData)

	// Call the LLM
	response, err := g.provider.Call(g.model, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate architecture: %w", err)
	}

	// Parse the response into an architecture
	architecture, err := g.parseArchitectureResponse(response.Content, interviewData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse architecture: %w", err)
	}

	architecture.ProjectID = interviewData.ProjectID
	architecture.CreatedAt = time.Now()

	return architecture, nil
}

// buildArchitecturePrompt creates the prompt for architecture generation
func (g *Generator) buildArchitecturePrompt(interviewData *state.InterviewData) string {
	prompt := `You are an expert software architect. Based on the following project requirements, generate a comprehensive system architecture.

PROJECT INFORMATION:
Problem Statement: ` + interviewData.ProblemStatement + `
Target Users: ` + strings.Join(interviewData.TargetUsers, ", ") + `
Success Metrics: ` + strings.Join(interviewData.SuccessMetrics, ", ") + `

Please provide a detailed architecture document with the following sections:

1. SYSTEM OVERVIEW
   - High-level description of the system
   - Key architectural decisions

2. COMPONENTS
   - List all major components (frontend, backend, database, cache, etc.)
   - For each component: name, type, purpose, technologies, dependencies

3. DATA FLOWS
   - Describe 2-3 key user journeys
   - For each: name, description, step-by-step flow

4. TECHNOLOGY RATIONALE
   - Explain why each technology was chosen
   - Consider: language, framework, database, infrastructure

5. SCALING STRATEGY
   - Horizontal scaling approach
   - Vertical scaling approach
   - Caching strategy
   - Load balancing
   - Database scaling

6. API CONTRACT
   - List 5-10 key REST endpoints (method, path, description)
   - List any WebSocket events if applicable

7. DATABASE SCHEMA
   - List main tables with columns
   - Describe key relationships

8. SECURITY APPROACH
   - Authentication method
   - Authorization strategy
   - Encryption approach
   - Audit logging

9. OBSERVABILITY STRATEGY
   - Logging approach
   - Metrics collection
   - Distributed tracing

10. DEPLOYMENT ARCHITECTURE
    - Development environment
    - Staging environment
    - Production environment

11. RISK ASSESSMENT
    - List 3-5 potential risks
    - For each: name, probability (low/medium/high/critical), impact (low/medium/high/critical), mitigation

12. ASSUMPTIONS AND UNKNOWNS
    - List key assumptions
    - List unknowns that need clarification

Format your response as structured text that can be parsed. Use clear section headers and consistent formatting.`

	return prompt
}

// parseArchitectureResponse parses the LLM response into an Architecture struct
func (g *Generator) parseArchitectureResponse(response string, interviewData *state.InterviewData) (*Architecture, error) {
	// This is a simplified parser. In production, you'd want more robust parsing
	architecture := &Architecture{
		SystemOverview:  extractSection(response, "SYSTEM OVERVIEW", "COMPONENTS"),
		Components:      []Component{},
		DataFlows:       []DataFlow{},
		TechRationale:   make(map[string]string),
		ScalingStrategy: ScalingPlan{},
		APIContract:     APISpec{},
		DatabaseSchema:  Schema{},
		SecurityApproach: SecurityPlan{
			Authentication: extractSection(response, "Authentication", "Authorization"),
			Authorization:  extractSection(response, "Authorization", "Encryption"),
			Encryption:     extractSection(response, "Encryption", "Audit"),
			Audit:          extractSection(response, "Audit", "OBSERVABILITY"),
		},
		Observability: ObservabilityPlan{
			Logging: extractSection(response, "Logging", "Metrics"),
			Metrics: extractSection(response, "Metrics", "Tracing"),
			Tracing: extractSection(response, "Tracing", "DEPLOYMENT"),
		},
		Deployment: DeploymentPlan{
			Development: extractSection(response, "Development", "Staging"),
			Staging:     extractSection(response, "Staging", "Production"),
			Production:  extractSection(response, "Production", "RISK"),
		},
		Risks:       []Risk{},
		Assumptions: []string{},
		Unknowns:    []string{},
	}

	// Extract components (simplified)
	componentsSection := extractSection(response, "COMPONENTS", "DATA FLOWS")
	if componentsSection != "" {
		architecture.Components = append(architecture.Components, Component{
			Name:         "Backend API",
			Type:         ComponentBackend,
			Purpose:      "Core business logic",
			Technologies: []string{},
			Dependencies: []string{},
		})
	}

	return architecture, nil
}

// extractSection extracts a section from the response between two markers
func extractSection(text, startMarker, endMarker string) string {
	startIdx := strings.Index(text, startMarker)
	if startIdx == -1 {
		return ""
	}

	endIdx := strings.Index(text[startIdx:], endMarker)
	if endIdx == -1 {
		return strings.TrimSpace(text[startIdx+len(startMarker):])
	}

	return strings.TrimSpace(text[startIdx+len(startMarker) : startIdx+endIdx])
}

// ExportMarkdown exports the architecture as markdown
func (g *Generator) ExportMarkdown(architecture *Architecture) (string, error) {
	var md strings.Builder

	md.WriteString("# System Architecture\n\n")
	md.WriteString(fmt.Sprintf("**Project ID:** %s\n", architecture.ProjectID))
	md.WriteString(fmt.Sprintf("**Generated:** %s\n\n", architecture.CreatedAt.Format("2006-01-02 15:04:05")))

	md.WriteString("## System Overview\n\n")
	md.WriteString(architecture.SystemOverview + "\n\n")

	md.WriteString("## Components\n\n")
	for _, comp := range architecture.Components {
		md.WriteString(fmt.Sprintf("### %s (%s)\n\n", comp.Name, comp.Type))
		md.WriteString(fmt.Sprintf("**Purpose:** %s\n\n", comp.Purpose))
		if len(comp.Technologies) > 0 {
			md.WriteString(fmt.Sprintf("**Technologies:** %s\n\n", strings.Join(comp.Technologies, ", ")))
		}
		if len(comp.Dependencies) > 0 {
			md.WriteString(fmt.Sprintf("**Dependencies:** %s\n\n", strings.Join(comp.Dependencies, ", ")))
		}
	}

	md.WriteString("## Security Approach\n\n")
	md.WriteString(fmt.Sprintf("**Authentication:** %s\n\n", architecture.SecurityApproach.Authentication))
	md.WriteString(fmt.Sprintf("**Authorization:** %s\n\n", architecture.SecurityApproach.Authorization))
	md.WriteString(fmt.Sprintf("**Encryption:** %s\n\n", architecture.SecurityApproach.Encryption))
	md.WriteString(fmt.Sprintf("**Audit:** %s\n\n", architecture.SecurityApproach.Audit))

	md.WriteString("## Risks\n\n")
	for _, risk := range architecture.Risks {
		md.WriteString(fmt.Sprintf("### %s\n\n", risk.Name))
		md.WriteString(fmt.Sprintf("- **Probability:** %s\n", risk.Probability))
		md.WriteString(fmt.Sprintf("- **Impact:** %s\n", risk.Impact))
		md.WriteString(fmt.Sprintf("- **Mitigation:** %s\n\n", risk.Mitigation))
	}

	if len(architecture.Assumptions) > 0 {
		md.WriteString("## Assumptions\n\n")
		for _, assumption := range architecture.Assumptions {
			md.WriteString(fmt.Sprintf("- %s\n", assumption))
		}
		md.WriteString("\n")
	}

	if len(architecture.Unknowns) > 0 {
		md.WriteString("## Unknowns\n\n")
		for _, unknown := range architecture.Unknowns {
			md.WriteString(fmt.Sprintf("- %s\n", unknown))
		}
		md.WriteString("\n")
	}

	return md.String(), nil
}

// ExportJSON exports the architecture as JSON
func (g *Generator) ExportJSON(architecture *Architecture) (string, error) {
	jsonData, err := json.MarshalIndent(architecture, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal architecture: %w", err)
	}
	return string(jsonData), nil
}

// ArchitectureIteration represents a refinement of the architecture
type ArchitectureIteration struct {
	Timestamp   time.Time
	Section     string
	OldValue    string
	NewValue    string
	Reason      string
}

// RefineArchitecture refines a specific section of the architecture
func (g *Generator) RefineArchitecture(architecture *Architecture, section string, refinementRequest string) (*Architecture, error) {
	if g.provider == nil {
		return nil, fmt.Errorf("provider is required for architecture refinement")
	}

	prompt := fmt.Sprintf(`You are refining a system architecture. The user wants to modify the following section:

SECTION: %s
CURRENT CONTENT:
%s

REFINEMENT REQUEST:
%s

Please provide the updated content for this section, maintaining consistency with the rest of the architecture.`, 
		section, g.getSectionContent(architecture, section), refinementRequest)

	response, err := g.provider.Call(g.model, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to refine architecture: %w", err)
	}

	// Update the architecture with the refined content
	updatedArch := g.updateArchitectureSection(architecture, section, response.Content)
	
	return updatedArch, nil
}

// getSectionContent retrieves the content of a specific section
func (g *Generator) getSectionContent(architecture *Architecture, section string) string {
	switch section {
	case "system_overview":
		return architecture.SystemOverview
	case "scaling_strategy":
		return fmt.Sprintf("Horizontal: %s\nVertical: %s\nCaching: %s\nLoad Balancing: %s\nDatabase: %s",
			architecture.ScalingStrategy.HorizontalScaling,
			architecture.ScalingStrategy.VerticalScaling,
			architecture.ScalingStrategy.Caching,
			architecture.ScalingStrategy.LoadBalancing,
			architecture.ScalingStrategy.DatabaseScaling)
	case "security":
		return fmt.Sprintf("Auth: %s\nAuthz: %s\nEncryption: %s\nAudit: %s",
			architecture.SecurityApproach.Authentication,
			architecture.SecurityApproach.Authorization,
			architecture.SecurityApproach.Encryption,
			architecture.SecurityApproach.Audit)
	case "observability":
		return fmt.Sprintf("Logging: %s\nMetrics: %s\nTracing: %s",
			architecture.Observability.Logging,
			architecture.Observability.Metrics,
			architecture.Observability.Tracing)
	case "deployment":
		return fmt.Sprintf("Dev: %s\nStaging: %s\nProd: %s",
			architecture.Deployment.Development,
			architecture.Deployment.Staging,
			architecture.Deployment.Production)
	default:
		return ""
	}
}

// updateArchitectureSection updates a specific section with new content
func (g *Generator) updateArchitectureSection(architecture *Architecture, section string, newContent string) *Architecture {
	updated := *architecture
	
	switch section {
	case "system_overview":
		updated.SystemOverview = newContent
	case "scaling_strategy":
		// Parse the new content and update scaling strategy
		updated.ScalingStrategy.HorizontalScaling = newContent
	case "security":
		// Parse and update security approach
		updated.SecurityApproach.Authentication = newContent
	case "observability":
		// Parse and update observability
		updated.Observability.Logging = newContent
	case "deployment":
		// Parse and update deployment
		updated.Deployment.Development = newContent
	}
	
	return &updated
}

// ListRefinableSection returns the sections that can be refined
func (g *Generator) ListRefinableSections() []string {
	return []string{
		"system_overview",
		"components",
		"technology_rationale",
		"scaling_strategy",
		"api_contract",
		"database_schema",
		"security",
		"observability",
		"deployment",
		"risks",
	}
}

// ValidateArchitecture checks if the architecture is complete and consistent
func (g *Generator) ValidateArchitecture(architecture *Architecture) (bool, []string) {
	var issues []string

	if architecture.SystemOverview == "" {
		issues = append(issues, "System overview is missing")
	}

	if len(architecture.Components) == 0 {
		issues = append(issues, "No components defined")
	}

	if architecture.SecurityApproach.Authentication == "" {
		issues = append(issues, "Authentication approach not defined")
	}

	if architecture.SecurityApproach.Authorization == "" {
		issues = append(issues, "Authorization approach not defined")
	}

	if len(architecture.Risks) == 0 {
		issues = append(issues, "No risks identified")
	}

	return len(issues) == 0, issues
}
