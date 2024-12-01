package del

import "bytes"

type Items []by

func (c Items) Len() no         { return len(c) }
func (c Items) Less(i, j no) bo { return bytes.Compare(c[i], c[j]) < 0 }
func (c Items) Swap(i, j no)    { c[i], c[j] = c[j], c[i] }
