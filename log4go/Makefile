include $(GOROOT)/src/Make.$(GOARCH)

TARG=elog
GOFILES=\
	elog.go\
	elog_config.go\
	elog_socklog.go\
	elog_filelog.go

include $(GOROOT)/src/Make.pkg

elogtest :
	gotest -benchmarks=.* > /tmp/gotest && cat /tmp/gotest | grep -v WARN | grep -v INFO
