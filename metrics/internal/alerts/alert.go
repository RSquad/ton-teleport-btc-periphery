package alerts

type Alert interface {
	Check(dataSource AlertDataSource) (int, error)
}
