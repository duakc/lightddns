package gendoc

import (
	"errors"
	"fmt"
	goast "go/ast"
	"maps"
	"strings"

	"github.com/duakc/mt"
)

var ErrMissingMetaValue = errors.New("missing meta value")

type BaseDocument struct {
	Name string
	Lang map[LangCode]string
}

func (o *BaseDocument) FromComment(comment *goast.CommentGroup,
	child func(name string, tokenizer *Tokenizer) error,
) error {
	if comment == nil {
		return nil
	}
	return o.FromToken(NewTokenizerString(comment.Text()), child)
}

func (o *BaseDocument) FromToken(token *Tokenizer,
	child func(name string, token *Tokenizer) error,
) error {
	lang := maps.Clone(LangMap)
	for token.NextMeta() {
		name := token.MetaName()
		switch {
		case strings.HasPrefix(name, "@LANG"):
			code := LangDefault
			if idx := strings.Index(name[len("@LANG"):], "."); idx >= 0 {
				code = LangCode(name[len("@LANG")+1+idx:])
			}
			if _, ok := lang[string(code)]; !ok {
				return fmt.Errorf("unregistered language or duplicated language declare")
			}

			token.NextMeta()
			langText := token.FragmentText()
			if langText == "" {
				return fmt.Errorf("%s: %w", name, ErrMissingMetaValue)
			}
			o.Lang[code] = langText
			delete(lang, string(code))
			continue
		}
		if err := child(name, token); err != nil {
			return err
		}
	}
	return nil
}

func identNames(ids []*goast.Ident) string {
	return strings.Join(mt.Map(ids, func(s *goast.Ident) string {
		return s.Name
	}), ", ")
}
