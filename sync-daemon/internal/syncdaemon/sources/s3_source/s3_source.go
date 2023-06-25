package s3_source

type SyncSourceS3 struct {
	bucketName string
	region     string
}

func NewSyncSourceS3(bucketName string, region string) *SyncSourceS3 {
	return &SyncSourceS3{
		bucketName: bucketName,
		region:     region,
	}
}

func (s SyncSourceS3) IterateFiles(prefix string, processor func(path string) error) error {
	return nil
}
