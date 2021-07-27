package fake

type diff map[string][]byte

type FakeDiffRepository struct {
	diffs map[string]diff
}

func NewFakeDiffRepository() *FakeDiffRepository {
	return &FakeDiffRepository{
		diffs: make(map[string]diff),
	}
}

func (r *FakeDiffRepository) SaveDataSide(ID string, side string, data []byte) error {
	d := r.diffs[ID]
	if d == nil {
		r.diffs[ID] = make(map[string][]byte, 2)
		d = r.diffs[ID]
	}
	d[side] = data
	return nil
}

func (r *FakeDiffRepository) GetDataSidesByID(ID string) (map[string][]byte, error) {
	return r.diffs[ID], nil
}
