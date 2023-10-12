package namedlist

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Person struct {
	GivenName string
	Surname   string
}

type Parent struct {
	Person
	Child Person
}

type GrandParent struct {
	Person
	Child Parent
}

type ParentPointToChild struct {
	Person
	Child *Person
}

type PersonWithTags struct {
	GivenName string `db:"first_name"`
	Surname   string `db:"last_name"`
}

type TaggedItem struct {
	Name  string `db:"item_name"`
	Value int    `db:"-"`
}

type ParentWithTags struct {
	PersonWithTags
	Possession TaggedItem
}

type ParentWithTagsAndEmbeddedTaggedItem struct {
	PersonWithTags
	TaggedItem
}

type WithUnalteredStructField struct {
	Name      string
	CreatedAt time.Time `db:".,asis"`
}

func Test_FromStruct(t *testing.T) {
	t.Run("handles a pointer to a struct", func(t *testing.T) {
		person := &Person{
			GivenName: "John",
			Surname:   "Doe",
		}

		expected := []any{
			sql.Named("given_name", "John"),
			sql.Named("surname", "Doe"),
		}

		namedList, err := New()
		assert.Nil(t, err)
		found, err := namedList.FromStruct(person)
		assert.Nil(t, err)
		assert.Equal(t, expected, found)
	})

	t.Run("handles top-level embedded struct", func(t *testing.T) {
		parent := Parent{
			Person: Person{
				GivenName: "John",
				Surname:   "Doe",
			},
			Child: Person{
				GivenName: "Jane",
				Surname:   "Doe",
			},
		}

		expected := []any{
			sql.Named("given_name", "John"),
			sql.Named("surname", "Doe"),
			sql.Named("child_given_name", "Jane"),
			sql.Named("child_surname", "Doe"),
		}

		namedList, _ := New()
		args, err := namedList.FromStruct(parent)
		assert.Nil(t, err)
		assert.NotEmpty(t, args)
		assert.Equal(t, expected, args)
	})

	t.Run("handles nested embedded struct", func(t *testing.T) {
		grandParent := GrandParent{
			Person: Person{
				GivenName: "John",
				Surname:   "Doe",
			},
			Child: Parent{
				Person: Person{
					GivenName: "Jane",
					Surname:   "Doe",
				},
				Child: Person{
					GivenName: "Jill",
					Surname:   "Eddy",
				},
			},
		}

		expected := []any{
			sql.Named("given_name", "John"),
			sql.Named("surname", "Doe"),
			sql.Named("child_given_name", "Jane"),
			sql.Named("child_surname", "Doe"),
			sql.Named("child_child_given_name", "Jill"),
			sql.Named("child_child_surname", "Eddy"),
		}

		namedList, _ := New()
		args, err := namedList.FromStruct(grandParent)
		assert.Nil(t, err)
		assert.NotEmpty(t, args)
		assert.Equal(t, expected, args)
	})

	t.Run("handles a pointer field", func(t *testing.T) {
		parent := ParentPointToChild{
			Person: Person{
				GivenName: "John",
				Surname:   "Doe",
			},
			Child: &Person{
				GivenName: "Jane",
				Surname:   "Doe",
			},
		}

		expected := []any{
			sql.Named("given_name", "John"),
			sql.Named("surname", "Doe"),
			sql.Named("child_given_name", "Jane"),
			sql.Named("child_surname", "Doe"),
		}

		namedList, _ := New()
		args, err := namedList.FromStruct(parent)
		assert.Nil(t, err)
		assert.NotEmpty(t, args)
		assert.Equal(t, expected, args)
	})

	t.Run("handles tags", func(t *testing.T) {
		parent := ParentWithTags{
			PersonWithTags: PersonWithTags{
				GivenName: "John",
				Surname:   "Doe",
			},
			Possession: TaggedItem{
				Name:  "Porsche 911",
				Value: 150_000,
			},
		}

		expected := []any{
			sql.Named("first_name", "John"),
			sql.Named("last_name", "Doe"),
			sql.Named("possession_item_name", "Porsche 911"),
		}

		namedList, _ := New()
		args, err := namedList.FromStruct(parent)
		assert.Nil(t, err)
		assert.NotEmpty(t, args)
		assert.Equal(t, expected, args)
	})

	t.Run("tags on embedded structs do not get prefixed", func(t *testing.T) {
		parent := ParentWithTagsAndEmbeddedTaggedItem{
			PersonWithTags: PersonWithTags{
				GivenName: "John",
				Surname:   "Doe",
			},
			TaggedItem: TaggedItem{
				Name:  "Porsche 911",
				Value: 150_000,
			},
		}

		expected := []any{
			sql.Named("first_name", "John"),
			sql.Named("last_name", "Doe"),
			sql.Named("item_name", "Porsche 911"),
		}

		namedList, _ := New()
		args, err := namedList.FromStruct(parent)
		assert.Nil(t, err)
		assert.NotEmpty(t, args)
		assert.Equal(t, expected, args)
	})

	t.Run("keeps snake name and recognizes asis tag", func(t *testing.T) {
		expectedTime := time.Time{}
		input := WithUnalteredStructField{
			Name:      "A Test",
			CreatedAt: expectedTime,
		}
		expected := []any{
			sql.Named("name", "A Test"),
			sql.Named("created_at", expectedTime),
		}

		namedList, _ := New()
		args, err := namedList.FromStruct(input)
		assert.Nil(t, err)
		assert.NotEmpty(t, args)
		assert.Equal(t, expected, args)
	})
}
