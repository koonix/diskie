package diskie

import (
	"cmp"
	"fmt"
	"slices"
)

type BlockMap struct {
	BlockMap map[string]*BlockDevice
}

func (bm *BlockMap) Filter(blocks []*BlockDevice, minImportance uint) ([]*BlockDevice, error) {
	if minImportance > 3 {
		return nil, fmt.Errorf("minImportance of %d is out of the possible range of 0 through 3", minImportance)
	}

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
		d := b.RootDrive
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

	return filtered, nil
}

func (bm *BlockMap) Sort() []*BlockDevice {

	blocks := make([]*BlockDevice, 0, len(bm.BlockMap))
	for _, v := range bm.BlockMap {
		blocks = append(blocks, v)
	}

	slices.SortStableFunc(blocks, func(a, b *BlockDevice) int {
		getSortKey := func(b *BlockDevice) string {
			sortKey := "000"
			d := b.RootDrive
			if d != nil && d.SortKey != nil {
				sortKey = *d.SortKey
			}

			rootSize := uint64(0)
			s := bm.BlockMap[b.RootDevice].PreferredSize
			if s != nil {
				rootSize = *s
			}

			usage := "00other"
			if b.IdUsage != nil {
				if *b.IdUsage == "filesystem" {
					usage = "02filesystem"
				} else if *b.IdUsage == "crypto" {
					usage = "01crypto"
				}
			}

			size := uint64(0)
			if b.PreferredSize != nil {
				size = *b.PreferredSize
			}

			return fmt.Sprintf("%s/%030d/%s/%030d", sortKey, rootSize, usage, size)
		}

		return cmp.Compare(getSortKey(b), getSortKey(a))
	})

	return blocks
}
