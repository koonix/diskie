package diskie

import (
	"cmp"
	"slices"
)

type BlockMap struct {
	BlockMap map[string]*BlockDevice
}

func (bm *BlockMap) Filter(blocks []*BlockDevice, minImportance uint) []*BlockDevice {
	filtered := make([]*BlockDevice, 0, len(blocks))

	importance := func(b *BlockDevice) uint {
		if b.IdUsage != nil && *b.IdUsage != "filesystem" && *b.IdUsage != "crypto" {
			return 0
		}
		if b.HintAuto != nil && *b.HintAuto {
			return 3
		}
		if b.HintIgnore != nil && *b.HintIgnore {
			return 1
		}
		if b.HintSystem != nil && *b.HintSystem {
			return 2
		}
		d := bm.getDrive(b)
		if d != nil && d.MediaAvailable != nil {
			if !*d.MediaAvailable {
				return 2
			}
		}
		return 1
	}

	for _, b := range blocks {
		if importance(b) >= minImportance {
			filtered = append(filtered, b)
		}
	}

	return filtered
}

func (bm *BlockMap) Sort() []*BlockDevice {

	blocks := make([]*BlockDevice, 0, len(bm.BlockMap))
	for _, v := range bm.BlockMap {
		blocks = append(blocks, v)
	}

	slices.SortStableFunc(blocks, func(a, b *BlockDevice) int {

		getSortKey := func(b *BlockDevice) string {
			d := bm.getDrive(b)
			if d != nil && d.SortKey != nil {
				return *d.SortKey
			}
			return "99"
		}

		getSize := func(block *BlockDevice) uint64 {
			if block.PreferredSize != nil {
				return *block.PreferredSize
			} else {
				return 0
			}
		}

		x := getSortKey(a)
		y := getSortKey(b)

		if x == y {
			return cmp.Compare(getSize(b), getSize(a))
		} else {
			return cmp.Compare(y, x)
		}
	})

	return blocks
}

func (bm *BlockMap) getDrive(block *BlockDevice) *Drive {
	if block.Drive != nil {
		return block.Drive
	}
	c := block.CryptoBackingDevice
	if c != nil && *c != "/" {
		b, has := bm.BlockMap[*c]
		if has {
			return bm.getDrive(b)
		}
	}
	return nil
}
