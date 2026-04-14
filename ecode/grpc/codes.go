package grpc

import (
	"context"

	"github.com/xslasd/goxf/ecode"
	pb "github.com/xslasd/goxf/ecode/grpc/proto"
)

type Codes struct{}

func (c Codes) GetCodes(ctx context.Context, empty *pb.Empty) (*pb.CodesRes, error) {
	mp := ecode.GetAllCodes()
	rows := make([]*pb.CodeInfo, 0)
	for k, v := range mp {
		if k < 1 {
			continue
		}
		row := &pb.CodeInfo{
			Code:    int64(k),
			Message: v,
		}
		rows = append(rows, row)
	}

	return &pb.CodesRes{
		Rows: rows,
	}, nil
}
