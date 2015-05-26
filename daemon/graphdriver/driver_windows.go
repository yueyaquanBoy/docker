package graphdriver

type DiffDiskDriver interface {
	Driver
	CopyDiff(id, sourceId string) error
}

func GetFSMagic(rootpath string) (FsMagic, error) {
	// Note it is OK to return FsMagicUnsupported on Windows.
	return FsMagicUnsupported, nil
}
