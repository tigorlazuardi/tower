package s3

type Object struct {
	Endpoint string
	Bucket   string
	Key      string
	Region   string
	Proto    string
}

func NewObject(proto, endpoint, region, bucket, key string) Object {
	return Object{
		Endpoint: endpoint,
		Bucket:   bucket,
		Key:      key,
		Region:   region,
		Proto:    proto,
	}
}
