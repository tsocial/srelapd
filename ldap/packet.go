package ldap

type SearchRequest struct {
	BaseDN       string
	Scope        int
	DerefAliases int
	SizeLimit    int
	TimeLimit    int
	TypesOnly    bool
	Filter       string
	Attributes   []string
	Controls     []Control
}

type Entry struct {
	DN         string
	Attributes []*EntryAttribute
}

type EntryAttribute struct {
	Name   string
	Values []string
}

const (
	ScopeBaseObject   = 0
	NeverDerefAliases = 0
	ScopeSingleLevel  = 1
	ScopeWholeSubtree = 2
)
