package main

import (
	"icelock/log"
	"os"
	"strings"

	"github.com/landlock-lsm/go-landlock/landlock"
	llsys "github.com/landlock-lsm/go-landlock/landlock/syscall"
)

const (
	// The set of access rights that apply to files
	accessFile landlock.AccessFSSet = llsys.AccessFSExecute | llsys.AccessFSWriteFile | llsys.AccessFSTruncate | llsys.AccessFSReadFile | llsys.AccessFSIoctlDev

	// The set of access rights associated with read access
	accessFSRead landlock.AccessFSSet = llsys.AccessFSReadFile | llsys.AccessFSReadDir

	// The set of access rights associated with write access
	accessFSWrite landlock.AccessFSSet = llsys.AccessFSWriteFile | llsys.AccessFSTruncate | llsys.AccessFSRemoveDir | llsys.AccessFSRemoveFile | llsys.AccessFSMakeChar | llsys.AccessFSMakeDir | llsys.AccessFSMakeReg | llsys.AccessFSMakeSock | llsys.AccessFSMakeFifo | llsys.AccessFSMakeBlock | llsys.AccessFSMakeSym | llsys.AccessFSRefer | llsys.AccessFSIoctlDev

	// The set of access rights associated with read/write access
	accessFSReadWrite landlock.AccessFSSet = accessFSRead | accessFSWrite

	// The set of filesystem access rights
	accessFSAll landlock.AccessFSSet = accessFSReadWrite | llsys.AccessFSExecute

	// The set of network access rights
	accessNetAll landlock.AccessNetSet = llsys.AccessNetBindTCP | llsys.AccessNetConnectTCP

	// The set of scoped access rights
	accessScopedAll landlock.ScopedSet = llsys.ScopeSignal | llsys.ScopeAbstractUnixSocket
)

var home string

func init() {
	home = os.Getenv("HOME")

	if home == "" {
		log.Error("$HOME is not set")
		os.Exit(1)
	}
}

func setupLandlock(cfg *config) {
	var rules []landlock.Rule

	var handledAccessFs landlock.AccessFSSet
	var handledAccessNet landlock.AccessNetSet
	var scopedAccess landlock.ScopedSet

	if cfg.FsRestricted {
		handledAccessFs = accessFSAll

		for _, path := range cfg.FsRO {
			rules = append(rules, roPath(path))
		}
		for _, path := range cfg.FsRX {
			rules = append(rules, rxPath(path))
		}
		for _, path := range cfg.FsRW {
			rules = append(rules, rwPath(path))
		}

		if len(cfg.FsRX) == 0 && !cfg.SeccompPrint {
			log.Error("Can't run the app, no executable paths were specified with the '--rx' flag")
		}
	}

	if cfg.NetRestricted {
		handledAccessNet = accessNetAll

		for _, port := range cfg.NetBindTCP {
			log.Debug("Allowing binding to TCP port %v", port)

			netRule := landlock.BindTCP(uint16(port))
			rules = append(rules, netRule)
		}
		for _, port := range cfg.NetConnectTCP {
			log.Debug("Allowing connecting to TCP port %v", port)

			netRule := landlock.ConnectTCP(uint16(port))
			rules = append(rules, netRule)
		}

	}

	if cfg.IpcScoped {
		scopedAccess = accessScopedAll
	}

	llConfig, err := landlock.NewConfig(handledAccessFs, handledAccessNet, scopedAccess)
	if err != nil {
		log.Error("Failed to create landlock config: %v", err)
		os.Exit(1)
	}
	// due to a bug in go-landlock this will report incorrect info, so currently
	// we use my fork that has a fix
	// TODO: switch back to upstream go-landlock once https://github.com/landlock-lsm/go-landlock/pull/49
	// is merged
	log.Info("Landlock config: %s", llConfig.String())

	err = llConfig.Restrict(rules...)
	if err != nil {
		log.Error("Failed to apply landlock restrictions: %v", err)
		os.Exit(1)
	}
	log.Info("Applied landlock restrictions")
}

func isDir(path string) bool {
	fileInfo, err := os.Stat(path)

	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

func expandTilde(path string) string {
	if strings.HasPrefix(path, "~") {
		return strings.Replace(path, "~", home, 1)
	}

	return path
}

func rxPath(path string) landlock.FSRule {
	expandedPath := expandTilde(path)
	log.Debug("Adding RX path %s", expandedPath)

	accessRights := accessFSRead&accessFile | llsys.AccessFSExecute

	if isDir(expandedPath) {
		accessRights = accessFSRead | llsys.AccessFSExecute
	}

	return landlock.PathAccess(accessRights, expandedPath).IgnoreIfMissing()
}

func roPath(path string) landlock.FSRule {
	expandedPath := expandTilde(path)
	log.Debug("Adding RO path %s", expandedPath)

	accessRights := accessFSRead & accessFile

	if isDir(expandedPath) {
		accessRights = accessFSRead
	}

	return landlock.PathAccess(accessRights, expandedPath).IgnoreIfMissing()
}

func rwPath(path string) landlock.FSRule {
	expandedPath := expandTilde(path)
	log.Debug("Adding RW path %s", expandedPath)

	accessRights := accessFSReadWrite & accessFile

	if isDir(expandedPath) {
		accessRights = accessFSReadWrite
	}

	return landlock.PathAccess(accessRights, expandedPath).IgnoreIfMissing()
}
