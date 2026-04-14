package ecode

import (
	"strconv"

	pb "github.com/xslasd/goxf/ecode/grpc/proto"
	"google.golang.org/protobuf/protoadapt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// A Status is an unsigned 32-bit error code as defined in the gRPC spec.
type Status uint32

const (
	Canceled Status = 1

	Unknown Status = 2

	InvalidArgument Status = 3

	DeadlineExceeded Status = 4

	NotFound Status = 5

	AlreadyExists Status = 6

	PermissionDenied Status = 7

	ResourceExhausted Status = 8

	FailedPrecondition Status = 9

	Aborted Status = 10

	OutOfRange Status = 11

	Unimplemented Status = 12

	Internal Status = 13

	Unavailable Status = 14

	DataLoss Status = 15

	Unauthenticated Status = 16

	_maxCode = 17
)

func GRPCError(e ECodes) error {
	return GRPCErrorWithCode(Unknown, e)
}

func GRPCErrorWithCode(s Status, e ECodes) error {
	dLen := len(e.Values())
	if dLen > 0 {
		details := &pb.CodeValues{Values: e.Values()}
		return GRPCErrorWithDetails(s, e, details)
	}
	return GRPCErrorWithDetails(s, e)
}

func GRPCErrorWithDetails(s Status, e ECodes, details ...protoadapt.MessageV1) error {
	stp := status.New(codes.Code(s), strconv.Itoa(e.Code()))
	st, err := stp.WithDetails(details...)
	if err != nil {
		return stp.Err()
	}
	return st.Err()
}

// FromGRPCError Functions may receive a GRPC error object as an input and then return a specific type of error object.
// Additional context or code snippets may be necessary for a complete understanding of the meaning and purpose of the function.
func FromGRPCError(err error) (Status, ECodes) {
	s, _ := status.FromError(err)
	if s != nil {
		switch Status(s.Code()) {
		case DeadlineExceeded:
			return Status(s.Code()), Deadline.WithDetails(s.Message())
		default:
			c, err := strconv.Atoi(s.Message())
			if err != nil {
				return Status(s.Code()), ServerErr.WithDetails(s.Message())
			}

			ds := s.Details()
			var details []any
			var f []string
			for index, detail := range ds {
				switch t := detail.(type) {
				case *pb.CodeValues:
					f = t.Values
					if len(ds) > 1 {
						details = append(ds[:index], ds[index+1:]...)
					}
				}
			}

			return Status(s.Code()), &ECode{id: c, msg: _codes[c], f: f, detail: details}
		}
	}
	return 0, OK
}
