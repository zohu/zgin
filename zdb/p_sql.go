package zdb

func LikeLeft(str string) string {
	return str + "%"
}
func LikeRight(str string) string {
	return "%" + str
}
func LikeBetween(str string) string {
	return "%" + str + "%"
}
