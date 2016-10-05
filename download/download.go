package download

import "github.com/alexandrasp/taskcluster-cli/extpoints"

//"github.com/taskcluster/httpbackoff"

func init() {
	extpoints.Register("download", download{})
}

type download struct{}

func (download) ConfigOptions() map[string]extpoints.ConfigOption {
	return nil
}

func (download) Summary() string {
	return "Download an artifact"
}

func (download) Usage() string {
	usage := "How you can download a taskcluster CLI artifact.\n"
	usage += "\n"
	usage += "Usage:\n"
	usage += "taskcluster download [options]"
	usage += "\n"
	usage += "Options:"
	usage += "\n"
	return usage
}

func (download) Execute(extpoints.Context) bool {

	return false
}
