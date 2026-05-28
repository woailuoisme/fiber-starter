package routing

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v3"
)

const Unassigned = "unassigned"

type Metadata struct {
	Feature    string
	Controller string
}

func NameForHandler(handler any) string {
	meta := MetadataFromFunctionName(functionName(handler))
	if meta.Feature == "" && meta.Controller == "" {
		return ""
	}

	target := targetFromFunctionName(functionName(handler))
	if target == "" {
		target = meta.Controller
	}

	if meta.Feature == "" {
		return target
	}
	if target == "" {
		return meta.Feature
	}
	return meta.Feature + ":" + target
}

func MetadataFromRoute(route fiber.Route) Metadata {
	if meta := MetadataFromName(route.Name); meta.Feature != "" || meta.Controller != "" {
		return meta
	}

	for idx := len(route.Handlers) - 1; idx >= 0; idx-- {
		meta := MetadataFromFunctionName(functionName(route.Handlers[idx]))
		if meta.Feature != "" || meta.Controller != "" {
			return meta
		}
	}

	return Metadata{
		Feature:    FeatureFromPath(route.Path),
		Controller: Unassigned,
	}
}

func MetadataFromName(name string) Metadata {
	name = strings.TrimSpace(name)
	if name == "" {
		return Metadata{}
	}

	feature, target, hasFeature := strings.Cut(name, ":")
	if !hasFeature {
		return Metadata{Controller: controllerFromQualifiedName(name)}
	}

	return Metadata{
		Feature:    normalizeGroup(feature),
		Controller: controllerFromQualifiedName(target),
	}
}

func MetadataFromFunctionName(name string) Metadata {
	name = strings.TrimSuffix(strings.TrimSpace(name), "-fm")
	if name == "" {
		return Metadata{}
	}

	return Metadata{
		Feature:    featureFromFunctionName(name),
		Controller: controllerFromFunctionName(name),
	}
}

func FeatureFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "root"
	}

	if len(parts) >= 3 && parts[0] == "api" && strings.HasPrefix(parts[1], "v") {
		return strings.TrimSuffix(parts[2], "s")
	}

	return parts[0]
}

func functionName(handler any) string {
	if handler == nil {
		return ""
	}

	value := reflect.ValueOf(handler)
	if value.Kind() != reflect.Func {
		return ""
	}

	fn := runtime.FuncForPC(value.Pointer())
	if fn == nil {
		return ""
	}
	return fn.Name()
}

func featureFromFunctionName(name string) string {
	const marker = "/internal/features/"

	idx := strings.Index(name, marker)
	if idx < 0 {
		return ""
	}

	remainder := name[idx+len(marker):]
	feature, _, _ := strings.Cut(remainder, "/")
	feature, _, _ = strings.Cut(feature, ".")
	return normalizeGroup(feature)
}

func controllerFromFunctionName(name string) string {
	if receiver := receiverName(name); receiver != "" {
		return receiver
	}

	symbol := symbolName(name)
	if isAnonymousFunction(symbol) {
		return ""
	}
	return controllerFromQualifiedName(symbol)
}

func targetFromFunctionName(name string) string {
	name = strings.TrimSuffix(strings.TrimSpace(name), "-fm")
	receiver := receiverName(name)
	action := actionName(name)
	if receiver != "" && action != "" {
		return receiver + "." + action
	}
	if receiver != "" {
		return receiver
	}
	if isAnonymousFunction(action) {
		return ""
	}
	return action
}

func receiverName(name string) string {
	end := strings.LastIndex(name, ").")
	if end < 0 {
		return ""
	}

	prefix := name[:end]
	if start := strings.LastIndex(prefix, "(*"); start >= 0 {
		return prefix[start+2:]
	}
	if start := strings.LastIndex(prefix, "."); start >= 0 {
		return strings.TrimPrefix(prefix[start+1:], "*")
	}
	return ""
}

func symbolName(name string) string {
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		return name[idx+1:]
	}
	return name
}

func actionName(name string) string {
	if idx := strings.LastIndex(name, ")."); idx >= 0 {
		return symbolName(name[idx+2:])
	}
	return symbolName(name)
}

func controllerFromQualifiedName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	controller, _, _ := strings.Cut(name, ".")
	return normalizeGroup(controller)
}

func isAnonymousFunction(name string) bool {
	return strings.HasPrefix(name, "func") || strings.Contains(name, ".func")
}

func normalizeGroup(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return value
}
