package native

func MustDefaultNatives() *Natives {
	ns, err := DefaultNatives()
	if err != nil {
		panic(err)
	}
	return ns
}

func DefaultNatives() (*Natives, error) {
	ns, err := Merge(Permissions, Precompiles)
	if err != nil {
		return nil, err
	}
	return ns, nil
}
