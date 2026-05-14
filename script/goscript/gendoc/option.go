package gendoc

import (
	"errors"
	goast "go/ast"
	"slices"
	"strings"

	"github.com/duakc/mt"
)

var ErrMissingMetaValue = errors.New("missing meta value")

type TokenNameHandler interface {
	FromTokenName(name string, token *Tokenizer) error
}

type BaseDocument struct {
	Name    string
	Lang    map[LangCode]string
	hasLANG bool
}

func (o *BaseDocument) FromComment(comment *goast.CommentGroup,
	tokenHandler TokenNameHandler,
) error {
	if comment == nil {
		return nil
	}
	o.Lang = make(map[LangCode]string)
	err := o.FromToken(NewTokenizerString(comment.Text()), tokenHandler)
	if err != nil {
		return err
	}
	return nil
}

func (o *BaseDocument) FromToken(token *Tokenizer,
	tokenHandler TokenNameHandler,
) error {
	for token.NextMeta() {
		name := token.MetaName()
		switch name {
		case "@LANG":

		//case strings.HasPrefix(name, "@LANG"):
		//	code := LangDefault
		//	if idx := strings.Index(name[len("@LANG"):], "."); idx >= 0 {
		//		code = LangCode(name[len("@LANG")+1+idx:])
		//	}
		//	if _, ok := lang[string(code)]; !ok {
		//		return fmt.Errorf("unregistered language or duplicated language declare")
		//	}
		//
		//	token.NextMeta()
		//	langText := token.FragmentText()
		//	if langText == "" {
		//		return fmt.Errorf("%s: %w", name, ErrMissingMetaValue)
		//	}
		//	o.Lang[code] = langText
		//	delete(lang, string(code))
		default:
			if err := tokenHandler.FromTokenName(name, token); err != nil {
				return err
			}
		}

	}
	return nil
}

func identNames(ids []*goast.Ident) string {
	return strings.Join(mt.Map(ids, func(s *goast.Ident) string {
		return s.Name
	}), ", ")
}

func docArray(s string) []string {
	vv := strings.Split(s, ",")
	return slices.DeleteFunc(vv, func(s string) bool {
		return len(s) == 0
	})
}
