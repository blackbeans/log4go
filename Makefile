include $(GOROOT)/src/Make.$(GOARCH)

TARG=log4go.googlecode.com/svn/stable
GOFILES=\
	log4go.go\
	config.go\
  termlog.go\
	socklog.go\
	filelog.go\
  wrapper.go

include $(GOROOT)/src/Make.pkg

elogtest :
	gotest -benchmarks=.* > /tmp/gotest && cat /tmp/gotest | grep -v WARN | grep -v INFO
