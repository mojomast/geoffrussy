package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mojomast/geoffrussy/internal/config"
	"github.com/mojomast/geoffrussy/internal/interview"
	"github.com/mojomast/geoffrussy/internal/provider"
	"github.com/mojomast/geoffrussy/internal/state"
	"github.com/spf13/cobra"
)

var (
	interviewResume bool
	interviewModel  string
)

var interviewCmd = &cobra.Command{
	Use:   "interview",
	Short: "Start or resume project interview",
	Long: `Start a new project interview or resume an existing one.
The interview gathers essential information about your project through
a structured five-phase process.`,
	RunE: runInterview,
}

func init() {
	interviewCmd.Flags().BoolVar(&interviewResume, "resume", false, "Resume existing interview")
	interviewCmd.Flags().StringVar(&interviewModel, "model", "", "Model to use for interview")
}

func runInterview(cmd *cobra.Command, args []string) error {
	fmt.Println("🎤 Starting Project Interview...")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()

	cfgMgr := config.NewManager()
	if err := cfgMgr.Load(nil); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	projectID := filepath.Base(cwd)

	// Use of same database location as init command
	dbPath := filepath.Join(cwd, ".geoffrussy", "state.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open state store: %w", err)
	}
	defer store.Close()

	_, err = store.GetProject(projectID)
	if err != nil {
		return fmt.Errorf("project not found. Please run 'geoffrussy init' first: %w", err)
	}

	providerName, modelName, err := getProviderAndModel(cfgMgr, "interview", interviewModel)
	if err != nil {
		fmt.Println("\n⚠️  Could not automatically select provider and model")
		fmt.Println("   Available options:")
		fmt.Println("   1. Run './geoffrussy config' to set up providers")
		fmt.Println("   2. Run './geoffrussy config --list-providers' to see available models")
		fmt.Println("   3. Use '--model <model-name>' flag to specify a model")
		return fmt.Errorf("failed to get provider and model: %w", err)
	}

	fmt.Printf("📦 Using Provider: %s\n", providerName)
	fmt.Printf("🤖 Using Model: %s\n", modelName)
	fmt.Println()

	bridge := provider.NewBridge()
	if err := setupProvider(bridge, cfgMgr, providerName); err != nil {
		return fmt.Errorf("failed to setup provider: %w", err)
	}

	prov, err := bridge.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	engine := interview.NewEngine(store, prov, modelName)

	var session *interview.InterviewSession

	if interviewResume {
		fmt.Println("🔄 Resuming interview from previous session...")
		session, err = engine.ResumeInterview(projectID)
		if err != nil {
			return fmt.Errorf("failed to resume interview: %w", err)
		}
	} else {
		fmt.Println("🆕 Starting new interview session...")
		session, err = engine.StartInterview(projectID)
		if err != nil {
			return fmt.Errorf("failed to start interview: %w", err)
		}
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		question, err := engine.GetNextQuestion(session)
		if err != nil {
			return fmt.Errorf("failed to get next question: %w", err)
		}

		if question == nil {
			complete, missing := engine.ValidateCompleteness(session)
			if complete {
				fmt.Println("════════════════════════════════════════════════════════")
				fmt.Println("✅ Interview completed successfully!")

				summary, err := engine.GenerateSummary(session)
				if err != nil {
					return fmt.Errorf("failed to generate summary: %w", err)
				}

				fmt.Println("\n📊 Interview Summary")
				fmt.Println("─────────────────────────────────────────────")
				fmt.Println(summary)

				fmt.Println("\n💡 Next steps:")
				fmt.Println("   Run 'geoffrussy design' to generate architecture")
				fmt.Println("   Run 'geoffrussy config' to update configuration")
			} else {
				fmt.Println("⚠️  Interview is incomplete. Missing required answers:")
				for _, m := range missing {
					fmt.Printf("   - %s\n", m)
				}
			}

			return nil
		}

		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Phase %s - Question %d\n", session.CurrentPhase, session.CurrentQuestion)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("\n%s\n\n", question.Text)

		fmt.Printf("Your answer (or 'help' for suggestions, 'back' to go back): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)

		if answer == "back" {
			fmt.Println("⏮️  Going to previous question...")
			continue
		}

		if answer == "help" {
			fmt.Println("\n💡 Suggestions:")
			fmt.Println("   - Be specific about your problem")
			fmt.Println("   - Mention your target users")
			fmt.Println("   - List key features you need")
			fmt.Println("   - Describe any constraints")
			fmt.Println()
			continue
		}

		if err := engine.RecordAnswer(session, question.ID, answer); err != nil {
			return fmt.Errorf("failed to record answer: %w", err)
		}

		if err := engine.SaveSession(session); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}

		fmt.Println("✅ Answer saved!")
	}
}
