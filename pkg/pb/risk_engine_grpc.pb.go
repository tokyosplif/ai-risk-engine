package pb

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	RiskEngineService_AnalyzeTransaction_FullMethodName = "/riskengine.RiskEngineService/AnalyzeTransaction"
)

type RiskEngineServiceClient interface {
	AnalyzeTransaction(ctx context.Context, in *AnalyzeRequest, opts ...grpc.CallOption) (*AnalyzeResponse, error)
}

type riskEngineServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRiskEngineServiceClient(cc grpc.ClientConnInterface) RiskEngineServiceClient {
	return &riskEngineServiceClient{cc}
}

func (c *riskEngineServiceClient) AnalyzeTransaction(ctx context.Context, in *AnalyzeRequest, opts ...grpc.CallOption) (*AnalyzeResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AnalyzeResponse)
	err := c.cc.Invoke(ctx, RiskEngineService_AnalyzeTransaction_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type RiskEngineServiceServer interface {
	AnalyzeTransaction(context.Context, *AnalyzeRequest) (*AnalyzeResponse, error)
	mustEmbedUnimplementedRiskEngineServiceServer()
}

type UnimplementedRiskEngineServiceServer struct{}

func (UnimplementedRiskEngineServiceServer) AnalyzeTransaction(context.Context, *AnalyzeRequest) (*AnalyzeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method AnalyzeTransaction not implemented")
}
func (UnimplementedRiskEngineServiceServer) mustEmbedUnimplementedRiskEngineServiceServer() {}
func (UnimplementedRiskEngineServiceServer) testEmbeddedByValue()                           {}

type UnsafeRiskEngineServiceServer interface {
	mustEmbedUnimplementedRiskEngineServiceServer()
}

func RegisterRiskEngineServiceServer(s grpc.ServiceRegistrar, srv RiskEngineServiceServer) {
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&RiskEngineService_ServiceDesc, srv)
}

func _RiskEngineService_AnalyzeTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AnalyzeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RiskEngineServiceServer).AnalyzeTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RiskEngineService_AnalyzeTransaction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RiskEngineServiceServer).AnalyzeTransaction(ctx, req.(*AnalyzeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var RiskEngineService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "riskengine.RiskEngineService",
	HandlerType: (*RiskEngineServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AnalyzeTransaction",
			Handler:    _RiskEngineService_AnalyzeTransaction_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/proto/risk_engine.proto",
}
