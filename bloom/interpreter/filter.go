package interpreter

type Filter struct{}

type FilterHandler interface{
	ProcessFilter(*Filter)
}
