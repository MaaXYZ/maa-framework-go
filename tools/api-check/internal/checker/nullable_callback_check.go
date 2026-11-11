package checker

import (
	"go/ast"
	"go/token"
	"strconv"
)

const (
	iosDeviceCancelCallbackName    = "MaaToolkitIOSDeviceCancelCallback"
	iosDeviceCancelCallbackPtrType = iosDeviceCancelCallbackName + "Ptr"
	iosDeviceCancelCallbackMarker  = "callback-ptr:" + iosDeviceCancelCallbackName
)

// The nullable iOS callback must bypass purego's automatic func conversion:
// nil funcs are rejected on Unix, and every conversion retains a trampoline.
// Keep this representation tied to one parameter, not to arbitrary C callbacks.
func nativeParamTypesMatch(module, function string, goTypes, cTypes []string) bool {
	if len(goTypes) != len(cTypes) {
		return false
	}
	for i, goType := range goTypes {
		if module == "toolkit" && function == "MaaToolkitIOSDeviceFindWithOptions" && i == 2 {
			if goType != iosDeviceCancelCallbackMarker || cTypes[i] != "callback:"+iosDeviceCancelCallbackName {
				return false
			}
			continue
		}
		if goType != cTypes[i] {
			return false
		}
	}
	return true
}

func checkIOSDeviceCancelCallbackPointerDeclaration(source *ast.File) []issue {
	for _, decl := range source.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec := spec.(*ast.TypeSpec)
			if typeSpec.Name.Name != iosDeviceCancelCallbackPtrType {
				continue
			}
			underlying, ok := typeSpec.Type.(*ast.Ident)
			if typeSpec.Assign.IsValid() || !ok || underlying.Name != "uintptr" {
				return []issue{{section: sectionController, message: "nullable callback pointer must be a defined uintptr type: " + iosDeviceCancelCallbackPtrType}}
			}
		}
	}
	return nil
}

func checkIOSDeviceCancelCallbackBinding(source *ast.File) []issue {
	importsPurego := false
	for _, spec := range source.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err == nil && path == "github.com/ebitengine/purego" && (spec.Name == nil || spec.Name.Name == "purego") {
			importsPurego = true
		}
	}
	bindings := 0
	for _, decl := range source.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			value := spec.(*ast.ValueSpec)
			if len(value.Names) != 1 || value.Names[0].Name != "iosDeviceCancelCallback" || len(value.Values) != 1 {
				continue
			}
			call, ok := value.Values[0].(*ast.CallExpr)
			if !ok || !isPuregoNewCallback(call) || len(call.Args) != 1 {
				continue
			}
			target, ok := call.Args[0].(*ast.Ident)
			if ok && target.Name == "_"+iosDeviceCancelCallbackName+"Agent" {
				bindings++
			}
		}
	}
	callbackCalls := 0
	ast.Inspect(source, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok && isPuregoNewCallback(call) {
			callbackCalls++
		}
		return true
	})
	if !importsPurego || bindings != 1 || callbackCalls != 1 {
		return []issue{{section: sectionController, message: "nullable callback binding mismatch for " + iosDeviceCancelCallbackName + ": require one shared purego.NewCallback(_" + iosDeviceCancelCallbackName + "Agent) registration in iosDeviceCancelCallback"}}
	}
	return nil
}

func isPuregoNewCallback(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "NewCallback" {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)
	return ok && pkg.Name == "purego"
}
