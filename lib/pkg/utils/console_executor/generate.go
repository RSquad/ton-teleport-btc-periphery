package console_executor

//go:generate sh -c "rm -rf mocks && mkdir -p mocks"
//go:generate minimock -i ConsoleExecutorInterface -o ./mocks/ -s "_minimock.go"
