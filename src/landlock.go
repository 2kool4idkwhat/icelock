package main

import (
	"icelock/log"
	"os"

	"github.com/landlock-lsm/go-landlock/landlock"
	llsys "github.com/landlock-lsm/go-landlock/landlock/syscall"
)

const (
	// The set of access rights that only apply to files
	accessFile landlock.AccessFSSet = llsys.AccessFSExecute | llsys.AccessFSWriteFile | llsys.AccessFSTruncate | llsys.AccessFSReadFile

	// The set of access rights associated with read access
	accessFSRead landlock.AccessFSSet = llsys.AccessFSReadFile | llsys.AccessFSReadDir

	// The set of access rights associated with write access
	accessFSWrite landlock.AccessFSSet = llsys.AccessFSWriteFile | llsys.AccessFSTruncate | llsys.AccessFSRemoveDir | llsys.AccessFSRemoveFile | llsys.AccessFSMakeChar | llsys.AccessFSMakeDir | llsys.AccessFSMakeReg | llsys.AccessFSMakeSock | llsys.AccessFSMakeFifo | llsys.AccessFSMakeBlock | llsys.AccessFSMakeSym | llsys.AccessFSRefer | llsys.AccessFSIoctlDev

	// The set of access rights associated with read/write access
	accessFSReadWrite landlock.AccessFSSet = accessFSRead | accessFSWrite
)

func setupLandlock(cfg *config) {
	var rules []landlock.Rule

	if cfg.FsRestricted {
		for _, path := range cfg.FsRO {
			rules = append(rules, roPath(path))
		}
		for _, path := range cfg.FsRX {
			rules = append(rules, rxPath(path))
		}
		for _, path := range cfg.FsRW {
			rules = append(rules, rwPath(path))
		}
	}

	if cfg.NetRestricted {
		for _, port := range cfg.NetBindTcp {
			log.Debug("Allowing binding to TCP port %v", port)

			netRule := landlock.BindTCP(uint16(port))
			rules = append(rules, netRule)
		}
		for _, port := range cfg.NetConnectTcp {
			log.Debug("Allowing connecting to TCP port %v", port)

			netRule := landlock.ConnectTCP(uint16(port))
			rules = append(rules, netRule)
		}

	}

	switch {
	case cfg.FsRestricted && cfg.NetRestricted:
		err := landlock.V5.Restrict(rules...)
		if err != nil {
			log.Error("Failed to apply landlock restrictions (filesystem + network): %v", err)
			os.Exit(1)
		}

		log.Info("Applied landlock restrictions (filesystem + network)")

	case cfg.FsRestricted && !cfg.NetRestricted:
		err := landlock.V5.RestrictPaths(rules...)
		if err != nil {
			log.Error("Failed to apply landlock filesystem restrictions: %v", err)
			os.Exit(1)
		}

		log.Info("Applied landlock filesystem restrictions")

	case !cfg.FsRestricted && cfg.NetRestricted:
		err := landlock.V5.RestrictNet(rules...)
		if err != nil {
			log.Error("Failed to apply landlock network restrictions: %v", err)
			os.Exit(1)
		}

		log.Info("Applied landlock network restrictions")

	default:
	}

}

func isDir(path string) bool {
	fileInfo, err := os.Stat(path)

	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

func rxPath(path string) landlock.FSRule {
	log.Debug("Adding RX path %s", path)

	accessRights := accessFSRead&accessFile | llsys.AccessFSExecute

	if isDir(path) {
		accessRights = accessFSRead | llsys.AccessFSExecute
	}

	return landlock.PathAccess(accessRights, path).IgnoreIfMissing()
}

func roPath(path string) landlock.FSRule {
	log.Debug("Adding RO path %s", path)

	accessRights := accessFSRead & accessFile

	if isDir(path) {
		accessRights = accessFSRead
	}

	return landlock.PathAccess(accessRights, path).IgnoreIfMissing()
}

func rwPath(path string) landlock.FSRule {
	log.Debug("Adding RW path %s", path)

	accessRights := accessFSReadWrite & accessFile

	if isDir(path) {
		accessRights = accessFSReadWrite
	}

	return landlock.PathAccess(accessRights, path).IgnoreIfMissing()
}

func setupLandlockIpc(cfg *config) {
	if cfg.IpcScoped {

		err := landlock.V6.RestrictScoped()
		if err != nil {
			log.Error("Failed to apply landlock IPC restrictions: %v", err)
			os.Exit(1)
		}

		log.Info("Applied landlock IPC restrictions")
	}
}
