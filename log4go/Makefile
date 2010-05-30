include $(GOROOT)/src/Make.$(GOARCH)

TARG=log4go
GOFILES=\
	log4go.go\
	log4go_config.go\
  log4go_termlog.go\
	log4go_socklog.go\
	log4go_filelog.go\
  log4go_wrapper.go

include $(GOROOT)/src/Make.pkg

elogtest :
	gotest -benchmarks=.* > /tmp/gotest && cat /tmp/gotest | grep -v WARN | grep -v INFO
