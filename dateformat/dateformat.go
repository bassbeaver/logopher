package dateformat

const (
	Y = "2006"
	M0 = "01"
	D0 = "02"
	H = "15"
	I = "04"
	S = "05"

	DateFormat = Y+"-"+M0+"-"+D0
	DateTimeFormat = DateFormat+" "+H+":"+I+":"+S
)