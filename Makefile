# Expected environment variables: $GOROOT

INSTALL_OPTS = -C
INSTALL_ROOT = $(GOROOT)/src/pkg/elog
SOURCE_FILES = src/Makefile src/elog.go src/elog_config.go src/elog_socklog.go src/elog_filelog.go src/elog_test.go

INSTALL = install
MAKE = gomake

install :
	$(INSTALL) -d $(INSTALL_ROOT)
	$(INSTALL) $(INSTALL_OPTS) $(SOURCE_FILES) $(INSTALL_ROOT)
	cd $(INSTALL_ROOT) && $(MAKE) install

test : install
	cd $(INSTALL_ROOT) && $(MAKE) test
