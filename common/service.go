package common 

import "context"

type Service interface{
    GetYTData(ctx context.Context) (any, error)
}
