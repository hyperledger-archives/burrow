package slice

func Slice(elements ...interface{}) []interface{} {
	return elements
}

func EmptySlice() []interface{} {
	return []interface{}{}
}

// Like append but on the interface{} type and always to a fresh backing array
// so can be used safely with slices over arrays you did not create.
func CopyAppend(slice []interface{}, elements ...interface{}) []interface{} {
	sliceLength := len(slice)
	newSlice := make([]interface{}, sliceLength+len(elements))
	for i, e := range slice {
		newSlice[i] = e
	}
	for i, e := range elements {
		newSlice[sliceLength+i] = e
	}
	return newSlice
}

// Prepend elements to slice in the order they appear
func CopyPrepend(slice []interface{}, elements ...interface{}) []interface{} {
	elementsLength := len(elements)
	newSlice := make([]interface{}, len(slice)+elementsLength)
	for i, e := range elements {
		newSlice[i] = e
	}
	for i, e := range slice {
		newSlice[elementsLength+i] = e
	}
	return newSlice
}

// Concatenate slices into a single slice
func Concat(slices ...[]interface{}) []interface{} {
	offset := 0
	for _, slice := range slices {
		offset += len(slice)
	}
	concat := make([]interface{}, offset)
	offset = 0
	for _, slice := range slices {
		for i, e := range slice {
			concat[offset+i] = e
		}
		offset += len(slice)
	}
	return concat
}

// Deletes n elements starting with the ith from a slice by splicing.
// Beware uses append so the underlying backing array will be modified!
func Delete(slice []interface{}, i int, n int) []interface{} {
	return append(slice[:i], slice[i+n:]...)
}

//
func DeleteAt(slice []interface{}, i int) []interface{} {
	return Delete(slice, i, 1)
}
