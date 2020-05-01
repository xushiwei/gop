package goprj

import (
	"go/ast"
	"go/token"
	"reflect"

	"github.com/qiniu/x/log"
)

// -----------------------------------------------------------------------------

// InferType infers type from a ast.Expr.
func (p *fileLoader) InferType(expr ast.Expr) Type {
	switch v := expr.(type) {
	case *ast.CallExpr:
		return p.inferTypeFromFun(v.Fun)
	case *ast.UnaryExpr:
		switch v.Op {
		case token.AND: // &X
			t := p.InferType(v.X)
			return p.prj.UniqueType(&PointerType{Elem: t})
		default:
			log.Fatalln("InferType: unknown UnaryExpr -", v.Op)
		}
	case *ast.SelectorExpr:
		t := p.InferType(v.X)
		if tf, ok := TypeField(t, v.Sel.Name); ok {
			return tf
		}
	case *ast.Ident:
	case *ast.CompositeLit:
		if v.Type != nil {
			return p.ToType(v.Type)
		}
		return p.inferCompositeLit(v.Elts)
	case *ast.ArrayType:
	}
	log.Fatalln("InferType: unknown -", reflect.TypeOf(expr), "-", expr)
	return nil
}

func (p *fileLoader) inferCompositeLit(elts []ast.Expr) Type {
	for _, elt := range elts {
		log.Fatalln("inferCompositeLit: unknown -", reflect.TypeOf(elt))
	}
	return nil
}

func (p *fileLoader) inferTypeFromFun(fun ast.Expr) Type {
	switch v := fun.(type) {
	case *ast.SelectorExpr:
		switch recv := v.X.(type) {
		case *ast.Ident:
			fnt, err := p.pkg.FindPackageType(recv.Name, v.Sel.Name)
			if err == nil {
				if fn, ok := fnt.(*FuncType); ok {
					return &RetType{Results: fn.Results}
				}
				log.Fatalln("inferTypeFromFun: FindPackageType not func -", reflect.TypeOf(fnt))
			}
			log.Fatalln("inferTypeFromFun: FindPackageType error -", err, "-", recv.Name, v.Sel)
		default:
			if trecv := p.InferType(recv); trecv != nil {
				if tf, ok := TypeField(trecv, v.Sel.Name); ok {
					if fn, ok := tf.(*FuncType); ok {
						return &RetType{Results: fn.Results}
					}
				}
			}
		}
	case *ast.Ident:
		return &UninferedRetType{Fun: v.Name, Nth: -1}
	}
	log.Fatalln("inferTypeFromFun: unknown -", reflect.TypeOf(fun))
	return nil
}

// -----------------------------------------------------------------------------