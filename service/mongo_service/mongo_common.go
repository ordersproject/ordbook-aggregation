package mongo_service

const (
	EQ_     = ""
	NOT_EQ_ = "$ne"     //≠
	LT_     = "$lt"     //<
	LTE_    = "$lte"    //<=
	GT_     = "$gt"     //>
	GTE_    = "$gte"    //>=
	IN_     = "$in"     //
	NOT_IN_ = "$nin"    //
	LIKE_   = "$regex"  //
	EXISTS_ = "$exists" //
	OR_     = "$or"     //
	AND_    = "$and"    //
)

type OrderByType int64

const (
	NoOrder_ OrderByType = 0
	ASC_     OrderByType = 1
	DESC_    OrderByType = -1
)