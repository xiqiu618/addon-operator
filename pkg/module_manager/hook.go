package module_manager

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/flant/addon-operator/pkg/utils"
	"github.com/kennygrant/sanitize"
	log "github.com/sirupsen/logrus"

	utils_file "github.com/flant/shell-operator/pkg/utils/file"
)

type Hook interface {
	WithModuleManager(moduleManager *MainModuleManager)
	WithConfig(configOutput []byte) (err error)
	GetName() string
	GetPath() string
	PrepareTmpFilesForHookRun(context interface{}) (map[string]string, error)
	Order(binding BindingType) float64
}

type CommonHook struct {
	// The unique name like 'global-hooks/startup_hook' or '002-module/hooks/cleanup'.
	Name           string
	// The absolute path of the executable file.
	Path           string

	moduleManager *MainModuleManager
}

func (c *CommonHook) WithModuleManager(moduleManager *MainModuleManager) {
	c.moduleManager = moduleManager
}

func (h *CommonHook) SafeName() string {
	return sanitize.BaseName(h.Name)
}

func (h *CommonHook) GetName() string {
	return h.Name
}

func (h *CommonHook) GetPath() string {
	return h.Path
}

// SearchGlobalHooks recursively find all executables in hooksDir. Absent hooksDir is not an error.
func SearchGlobalHooks(hooksDir string) (hooks []*GlobalHook, err error) {
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return nil, nil
	}

	hooksRelativePaths, err := utils_file.RecursiveGetExecutablePaths(hooksDir)
	if err != nil {
		return nil, err
	}

	hooks = make([]*GlobalHook, 0)

	// sort hooks by path
	sort.Strings(hooksRelativePaths)
	log.Debugf("  Hook paths: %+v", hooksRelativePaths)

	for _, hookPath := range hooksRelativePaths {
		hookName, err := filepath.Rel(hooksDir, hookPath)
		if err != nil {
			return nil, err
		}

		globalHook := NewGlobalHook(hookName, hookPath)

		hooks = append(hooks, globalHook)
	}

	return
}

func SearchModuleHooks(module *Module) (hooks []*ModuleHook, err error) {
	hooksDir := filepath.Join(module.Path, "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return nil, nil
	}

	hooksRelativePaths, err := utils_file.RecursiveGetExecutablePaths(hooksDir)
	if err != nil {
		return nil, err
	}

	hooks = make([]*ModuleHook, 0)

	// sort hooks by path
	sort.Strings(hooksRelativePaths)
	log.Debugf("  Hook paths: %+v", hooksRelativePaths)

	for _, hookPath := range hooksRelativePaths {
		hookName, err := filepath.Rel(filepath.Dir(module.Path), hookPath)
		if err != nil {
			return nil, err
		}

		moduleHook := NewModuleHook(hookName, hookPath)
		moduleHook.WithModule(module)

		hooks = append(hooks, moduleHook)
	}

	return
}


func (mm *MainModuleManager) RegisterGlobalHooks() error {
	log.Debug("Search and register global hooks")

	mm.globalHooksOrder = make(map[BindingType][]*GlobalHook)
	mm.globalHooksByName = make(map[string]*GlobalHook)

	hooks, err := SearchGlobalHooks(mm.GlobalHooksDir)
	if err != nil {
		return err
	}
	log.Debug("Found %d global hooks", len(hooks))

	for _, globalHook := range hooks {
		logEntry := log.WithField("hook", globalHook.Name).
			WithField("hook.type", "global")

		configOutput, err := NewHookExecutor(globalHook, nil).Config()
		if err != nil {
			logEntry.Errorf("Run --config: %s", err)
			return fmt.Errorf("global hook --config run problem")
		}

		err = globalHook.WithConfig(configOutput)
		if err != nil {
			logEntry.Errorf("Hook return bad config: %s", err)
			return fmt.Errorf("global hook return bad config")
		}

		globalHook.WithModuleManager(mm)
		// register global hook in indexes
		for _, binding := range globalHook.Config.Bindings() {
			mm.globalHooksOrder[binding] = append(mm.globalHooksOrder[binding], globalHook)
		}
		mm.globalHooksByName[globalHook.Name] = globalHook

		logEntry.Infof("Registered")
	}

	return nil
}

func (mm *MainModuleManager) RegisterModuleHooks(module *Module, logLabels map[string]string) error {
	logEntry := log.WithFields(utils.LabelsToLogFields(logLabels)).WithField("module", module.Name)

	if _, ok := mm.modulesHooksOrderByName[module.Name]; ok {
		logEntry.Debugf("Module hooks already registered")
		return nil
	}

	logEntry.Debugf("Search and register hooks")

	hooks, err := SearchModuleHooks(module)
	if err != nil {
		return err
	}
	logEntry.Debugf("Found %d hooks", len(hooks))

	for _, moduleHook := range hooks {
		hookLogEntry := logEntry.WithField("hook", moduleHook.Name).
			WithField("hook.type", "module")

		configOutput, err := NewHookExecutor(moduleHook, nil).Config()
		if err != nil {
			hookLogEntry.Errorf("Run --config: %s", err)
			return fmt.Errorf("module hook --config run problem")
		}

		err = moduleHook.WithConfig(configOutput)
		if err != nil {
			hookLogEntry.Errorf("Hook return bad config: %s", err)
			return fmt.Errorf("module hook return bad config")
		}

		moduleHook.WithModuleManager(mm)
		// register module hook in indexes
		for _, binding := range moduleHook.Config.Bindings() {
			if mm.modulesHooksOrderByName[module.Name] == nil {
				mm.modulesHooksOrderByName[module.Name] = make(map[BindingType][]*ModuleHook)
			}
			mm.modulesHooksOrderByName[module.Name][binding] = append(mm.modulesHooksOrderByName[module.Name][binding], moduleHook)
		}

		hookLogEntry.Infof("Registered")
	}

	return nil
}
