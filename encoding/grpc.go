package encoding

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCodec(&GRPCCodec{})
}

const GRPCCodecName = "gogo"

type GRPCCodec struct {
}

func (G *GRPCCodec) String() string {
	return GRPCCodecName
}

func (G *GRPCCodec) Marshal(v interface{}) ([]byte, error) {
	return Encode(v.(proto.Message))
}

func (G *GRPCCodec) Unmarshal(data []byte, v interface{}) error {
	return Decode(data, v.(proto.Message))
}

func (G *GRPCCodec) Name() string {
	return GRPCCodecName
}

var DefaultDialOptions = []grpc.DialOption{grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.CallContentSubtype(GRPCCodecName))}

func GRPCDial(grpcAddress string, additionalOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(grpcAddress, append(additionalOpts, DefaultDialOptions...)...)
}

func GRPCDialContext(ctx context.Context, grpcAddress string, additionalOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, grpcAddress, append(additionalOpts, DefaultDialOptions...)...)
}
