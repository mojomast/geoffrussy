package design

import (
	"testing"
	"time"

	"github.com/mojomast/geoffrussy/internal/provider"
	"github.com/mojomast/geoffrussy/internal/state"
)

// MockProvider for testing
type MockProvider struct {
	response string
}

func (m *MockProvider) Name() string {
	return "mock"
}

func (m *MockProvider) Authenticate(apiKey string) error {
	return nil
}

func (m *MockProvider) IsAuthenticated() bool {
	return true
}

func (m *MockProvider) ListModels() ([]provider.Model, error) {
	return []provider.Model{}, nil
}

func (m *MockProvider) DiscoverModels() ([]provider.Model, error) {
	return []provider.Model{}, nil
}

func (m *MockProvider) Call(model string, prompt string) (*provider.Response, error) {
	return &provider.Response{
		Content:      m.response,
		TokensInput:  100,
		TokensOutput: 200,
		Model:        model,
		Provider:     "mock",
	}, nil
}

func (m *MockProvider) Stream(model string, prompt string) (<-chan string, error) {
	ch := make(chan string, 1)
	ch <- m.response
	close(ch)
	return ch, nil
}

func (m *MockProvider) GetRateLimitInfo() (*provider.RateLimitInfo, error) {
	return nil, nil
}

func (m *MockProvider) GetQuotaInfo() (*provider.QuotaInfo, error) {
	return nil, nil
}

func (m *MockProvider) SupportsCodingPlan() bool {
	return false
}

func TestDesignGenerator(t *testing.T) {
	mockResponse := `
SYSTEM OVERVIEW
This is a task management system with real-time collaboration features.

COMPONENTS
Backend API - REST API server
Frontend - React web application
Database - PostgreSQL for data storage

DATA FLOWS
User creates task: User -> Frontend -> Backend -> Database

TECHNOLOGY RATIONALE
Go for backend: Performance and concurrency
React for frontend: Component-based architecture
PostgreSQL: ACID compliance and reliability

SCALING STRATEGY
Horizontal scaling: Multiple API servers behind load balancer
Vertical scaling: Increase server resources as needed
Caching: Redis for session and frequently accessed data
Load balancing: Nginx reverse proxy
Database scaling: Read replicas for queries

API CONTRACT
POST /api/tasks - Create new task
GET /api/tasks - List all tasks
GET /api/tasks/:id - Get task details

DATABASE SCHEMA
tasks table: id, title, description, status, created_at
users table: id, email, name, created_at

SECURITY APPROACH
Authentication: JWT tokens
Authorization: Role-based access control
Encryption: TLS for transport, bcrypt for passwords
Audit: Log all data modifications

OBSERVABILITY STRATEGY
Logging: Structured JSON logs
Metrics: Prometheus metrics
Tracing: OpenTelemetry distributed tracing

DEPLOYMENT ARCHITECTURE
Development: Local Docker Compose
Staging: Kubernetes cluster with staging namespace
Production: Kubernetes cluster with production namespace

RISK ASSESSMENT
Database failure - high probability, high impact - Implement automated backups
API overload - medium probability, medium impact - Add rate limiting

ASSUMPTIONS AND UNKNOWNS
Assumptions: Users have modern browsers
Unknowns: Exact peak load requirements
`

	mockProvider := &MockProvider{response: mockResponse}
	generator := NewGenerator(mockProvider, "test-model")

	interviewData := &state.InterviewData{
		ProjectID:        "test-project",
		ProjectName:      "Test Project",
		ProblemStatement: "Need a task management system",
		TargetUsers:      []string{"Developers"},
		SuccessMetrics:   []string{"User engagement"},
		CreatedAt:        time.Now(),
	}

	t.Run("GenerateArchitecture", func(t *testing.T) {
		architecture, err := generator.GenerateArchitecture(interviewData)
		if err != nil {
			t.Fatalf("Failed to generate architecture: %v", err)
		}

		if architecture == nil {
			t.Fatal("Architecture should not be nil")
		}

		if architecture.ProjectID != "test-project" {
			t.Errorf("Expected project ID 'test-project', got '%s'", architecture.ProjectID)
		}

		if architecture.SystemOverview == "" {
			t.Error("System overview should not be empty")
		}
	})

	t.Run("ExportMarkdown", func(t *testing.T) {
		architecture := &Architecture{
			ProjectID:      "test-project",
			SystemOverview: "Test system",
			Components: []Component{
				{
					Name:         "Backend",
					Type:         ComponentBackend,
					Purpose:      "API server",
					Technologies: []string{"Go"},
					Dependencies: []string{"Database"},
				},
			},
			SecurityApproach: SecurityPlan{
				Authentication: "JWT",
				Authorization:  "RBAC",
				Encryption:     "TLS",
				Audit:          "Logs",
			},
			Risks: []Risk{
				{
					Name:        "Data loss",
					Probability: RiskMedium,
					Impact:      RiskHigh,
					Mitigation:  "Backups",
				},
			},
			Assumptions: []string{"Users have internet"},
			Unknowns:    []string{"Peak load"},
			CreatedAt:   time.Now(),
		}

		markdown, err := generator.ExportMarkdown(architecture)
		if err != nil {
			t.Fatalf("Failed to export markdown: %v", err)
		}

		if markdown == "" {
			t.Fatal("Markdown should not be empty")
		}

		// Check for key sections
		if !contains(markdown, "System Architecture") {
			t.Error("Markdown should contain 'System Architecture'")
		}

		if !contains(markdown, "Backend") {
			t.Error("Markdown should contain component name")
		}

		if !contains(markdown, "Security Approach") {
			t.Error("Markdown should contain security section")
		}

		if !contains(markdown, "Risks") {
			t.Error("Markdown should contain risks section")
		}
	})

	t.Run("ExportJSON", func(t *testing.T) {
		architecture := &Architecture{
			ProjectID:      "test-project",
			SystemOverview: "Test system",
			Components:     []Component{},
			CreatedAt:      time.Now(),
		}

		jsonStr, err := generator.ExportJSON(architecture)
		if err != nil {
			t.Fatalf("Failed to export JSON: %v", err)
		}

		if jsonStr == "" {
			t.Fatal("JSON should not be empty")
		}

		if !contains(jsonStr, "test-project") {
			t.Error("JSON should contain project ID")
		}
	})

	t.Run("GenerateArchitecture_NoProvider", func(t *testing.T) {
		generator := NewGenerator(nil, "test-model")

		_, err := generator.GenerateArchitecture(interviewData)
		if err == nil {
			t.Error("Should error when provider is nil")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsHelper(s, substr)
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestDesignGenerator_Reiteration(t *testing.T) {
	mockProvider := &MockProvider{response: "Updated system overview with better details"}
	generator := NewGenerator(mockProvider, "test-model")

	architecture := &Architecture{
		ProjectID:      "test-project",
		SystemOverview: "Original system overview",
		Components: []Component{
			{Name: "Backend", Type: ComponentBackend},
		},
		SecurityApproach: SecurityPlan{
			Authentication: "JWT",
			Authorization:  "RBAC",
		},
		Risks: []Risk{
			{Name: "Test risk", Probability: RiskLow, Impact: RiskLow},
		},
		CreatedAt: time.Now(),
	}

	t.Run("RefineArchitecture", func(t *testing.T) {
		refined, err := generator.RefineArchitecture(architecture, "system_overview", "Make it more detailed")
		if err != nil {
			t.Fatalf("Failed to refine architecture: %v", err)
		}

		if refined == nil {
			t.Fatal("Refined architecture should not be nil")
		}

		if refined.SystemOverview == architecture.SystemOverview {
			t.Error("System overview should have been updated")
		}
	})

	t.Run("ListRefinableSections", func(t *testing.T) {
		sections := generator.ListRefinableSections()

		if len(sections) == 0 {
			t.Error("Should have refinable sections")
		}

		// Check for key sections
		hasSystemOverview := false
		hasSecurity := false
		for _, section := range sections {
			if section == "system_overview" {
				hasSystemOverview = true
			}
			if section == "security" {
				hasSecurity = true
			}
		}

		if !hasSystemOverview {
			t.Error("Should include system_overview in refinable sections")
		}

		if !hasSecurity {
			t.Error("Should include security in refinable sections")
		}
	})

	t.Run("ValidateArchitecture_Valid", func(t *testing.T) {
		isValid, issues := generator.ValidateArchitecture(architecture)

		if !isValid {
			t.Errorf("Architecture should be valid, issues: %v", issues)
		}

		if len(issues) != 0 {
			t.Errorf("Should have no issues, got: %v", issues)
		}
	})

	t.Run("ValidateArchitecture_Invalid", func(t *testing.T) {
		invalidArch := &Architecture{
			ProjectID:      "test-project",
			SystemOverview: "",
			Components:     []Component{},
			SecurityApproach: SecurityPlan{
				Authentication: "",
				Authorization:  "",
			},
			Risks:     []Risk{},
			CreatedAt: time.Now(),
		}

		isValid, issues := generator.ValidateArchitecture(invalidArch)

		if isValid {
			t.Error("Architecture should be invalid")
		}

		if len(issues) == 0 {
			t.Error("Should have validation issues")
		}

		// Check for specific issues
		hasSystemOverviewIssue := false
		hasComponentsIssue := false
		for _, issue := range issues {
			if contains(issue, "System overview") {
				hasSystemOverviewIssue = true
			}
			if contains(issue, "components") {
				hasComponentsIssue = true
			}
		}

		if !hasSystemOverviewIssue {
			t.Error("Should identify missing system overview")
		}

		if !hasComponentsIssue {
			t.Error("Should identify missing components")
		}
	})

	t.Run("RefineArchitecture_NoProvider", func(t *testing.T) {
		generator := NewGenerator(nil, "test-model")

		_, err := generator.RefineArchitecture(architecture, "system_overview", "Update it")
		if err == nil {
			t.Error("Should error when provider is nil")
		}
	})
}
