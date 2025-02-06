package gadb

type (
	Feature  string
	Features map[Feature]struct{}
)

var (
	FeatSendrecvV2Brotli          = Feature("sendrecv_v2_brotli")
	FeatRemountShell              = Feature("remount_shell")
	FeatSendrecvV2                = Feature("sendrecv_v2")
	FeatAbbExec                   = Feature("abb_exec")
	FeatFixedPushMkdir            = Feature("fixed_push_mkdir")
	FeatFixedPushSymlinkTimestamp = Feature("fixed_push_symlink_timestamp")
	FeatAbb                       = Feature("abb")
	FeatShellV2                   = Feature("shell_v2")
	FeatCmd                       = Feature("cmd")
	FeatLsV2                      = Feature("ls_v2")
	FeatApex                      = Feature("apex")
	FeatStatV2                    = Feature("stat_v2")
)

func (fs Features) HasFeature(name Feature) bool {
	_, has := fs[name]
	return has
}
