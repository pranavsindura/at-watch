package publicRouter

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	envConstants "github.com/pranavsindura/at-watch/constants/env"
)

func Test(ctx *gin.Context) {
	html := `
<html>
	<head>
		<script src="http://code.jquery.com/jquery-1.8.3.js"></script>
		<script>
		</script>
	</head>
	<body>
		<!--A button to initiate a buy--!>
		<fyers-button 
			data-fyers="` + os.Getenv(envConstants.FyersAppID) + `" 
			data-symbol="NSE:SBIN-EQ" 
			data-product="CNC" 
			data-quantity="1" 
			data-price="102" 
			data-order_type="MARKET" 
			data-transaction_type="BUY" 
		>
		</fyers-button>
		<script src="https://api-connect-docs.fyers.in/fyers-lib.js"></script>
	</body>
</html>
`
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// https://api-connect.fyers.in/redirection
// <input name="data" type="hidden" value="[{&quot;symbol&quot;:&quot;NSE:RELIANCE-EQ&quot;,&quot;quantity&quot;:4,&quot;order_type&quot;:&quot;LIMIT&quot;,&quot;transaction_type&quot;:&quot;BUY&quot;,&quot;product&quot;:&quot;INTRADAY&quot;,&quot;disclosed_quantity&quot;:0,&quot;price&quot;:200}]">
// <input name="api_key" type="hidden" value="2S2A5JAL6L-100">

// https://api.fyers.in/api/v2/generate-authcode?client_id=XNQ201Q7YA-101&redirect_uri=https://api-connect.fyers.in/order?session_id=Trans_fyers75a57410fe0e42c8b2b95016b5a242d4&response_type=code&code_challenge=22fe491ae20e1d05c9c00047c85e72697edcf63114d6618d553ab55577d9a872&state=sample_state
// https://api.fyers.in/api/v2/generate-authcode?client_id=2S2A5JAL6L-100&redirect_uri=https://api-connect.fyers.in/order?session_id=Trans_fyers323659677d1b4339b8d9d549504a25c8&response_type=code&code_challenge=22fe491ae20e1d05c9c00047c85e72697edcf63114d6618d553ab55577d9a872&state=sample_state
