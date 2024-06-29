GO=go

main: float_converter

float_converter: float_converter.go
	GOPATH=`pwd` $(GO) build float_converter.go

clean:
	rm -f float_converter
	rm -f *.o
	rm -f *~
	rm -f \#*
