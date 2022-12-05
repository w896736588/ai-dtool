package helper

// FilterEmptyString 过滤掉空字符串数组
// @auth frog
// @date 2022-12-02 18:45:03
// @param arrayList
// @return []string
func FilterEmptyString(arrayList *[]string) []string {
	newList := make([]string, 0)
	for _, value := range *arrayList {
		if value != `` {
			newList = append(newList, value)
		}
	}
	return newList
}
