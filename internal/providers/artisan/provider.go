package artisan

import (
	"bytes"
	"errors"
	"strings"
	"sync"

	"lfiber/internal/providers/artisan/contracts"

	"github.com/spf13/cobra"
)

var (
	ErrCommandFactoryNotConfigured = errors.New("artisan command factory not configured")

	defaultFactoryMu sync.RWMutex
	defaultFactory   contracts.CommandFactory
)

// SetCommandFactory registers the application console command factory.
func SetCommandFactory(factory contracts.CommandFactory) {
	defaultFactoryMu.Lock()
	defer defaultFactoryMu.Unlock()
	defaultFactory = factory
}

// DefaultCommandFactory returns the registered application console command factory.
func DefaultCommandFactory() contracts.CommandFactory {
	defaultFactoryMu.RLock()
	defer defaultFactoryMu.RUnlock()
	return defaultFactory
}

// Register initializes and returns the Artisan command runner.
func Register(factory ...contracts.CommandFactory) (contracts.Artisan, error) {
	var resolved contracts.CommandFactory
	if len(factory) > 0 {
		resolved = factory[0]
	}
	return &Service{factory: resolved}, nil
}

type Service struct {
	factory contracts.CommandFactory
}

func (s *Service) Call(command string, args ...string) (contracts.Result, error) {
	root, err := s.rootCommand()
	if err != nil {
		return contracts.Result{ExitCode: 1}, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(append([]string{command}, args...))

	execErr := root.Execute()
	result := contracts.Result{
		ExitCode:    exitCode(execErr),
		Output:      stdout.String(),
		ErrorOutput: stderr.String(),
	}
	return result, execErr
}

func (s *Service) List() []contracts.CommandInfo {
	root, err := s.rootCommand()
	if err != nil {
		return nil
	}

	commands := make([]contracts.CommandInfo, 0)
	for _, cmd := range root.Commands() {
		collectCommands(root.CommandPath(), cmd, &commands)
	}
	return commands
}

func (s *Service) Has(command string) bool {
	command = strings.TrimSpace(command)
	if command == "" {
		return false
	}

	for _, info := range s.List() {
		if info.Available && info.Name == command {
			return true
		}
	}
	return false
}

func (s *Service) rootCommand() (*cobra.Command, error) {
	factory := s.factory
	if factory == nil {
		factory = DefaultCommandFactory()
	}
	if factory == nil {
		return nil, ErrCommandFactoryNotConfigured
	}

	root := factory()
	if root == nil {
		return nil, ErrCommandFactoryNotConfigured
	}
	return root, nil
}

func collectCommands(rootPath string, cmd *cobra.Command, commands *[]contracts.CommandInfo) {
	if cmd.IsAvailableCommand() {
		*commands = append(*commands, contracts.CommandInfo{
			Name:        commandName(rootPath, cmd),
			Description: cmd.Short,
			GroupID:     cmd.GroupID,
			Available:   true,
		})
	}
	for _, child := range cmd.Commands() {
		collectCommands(rootPath, child, commands)
	}
}

func commandName(rootPath string, cmd *cobra.Command) string {
	name := strings.TrimSpace(cmd.CommandPath())
	prefix := strings.TrimSpace(rootPath)
	if prefix != "" {
		name = strings.TrimSpace(strings.TrimPrefix(name, prefix))
	}
	return name
}

func exitCode(err error) int {
	if err != nil {
		return 1
	}
	return 0
}
