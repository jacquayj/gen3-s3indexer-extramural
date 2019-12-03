package common

type ManifestOpts struct {
	Regexs    []string `short:"r" long:"regex" description:"Object keys must match this or be skipped, multiple expressions can be specified" json:"regexs"`
	Prefix    *string  `short:"p" long:"prefix" description:"Limits the response to keys that begin with the specified prefix" json:"prefix"`
	BatchSize int      `short:"s" long:"batch-size" description:"Batch cluster size" default:"10" json:"batch_size"`
}

type BatchRun struct {
	StartKey *string `json:"start_key"`
	EndKey   *string `json:"end_key"`
}

type BatchRunRaw struct {
	StartKeyLine *int
	EndKeyLine   *int
}

type Jobs struct {
	BatchRuns    []BatchRun    `json:"jobs"`
	RawBatchRuns []BatchRunRaw `json:"-"`
	Opts         ManifestOpts  `json:"opts"`
	ObjCount     int           `json:"obj_count"`
}
