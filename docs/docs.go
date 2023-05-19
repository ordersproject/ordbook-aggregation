// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "API Support"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/brc20/order/push": {
            "post": {
                "description": "Push order",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "brc20"
                ],
                "summary": "Push order",
                "parameters": [
                    {
                        "description": "Request",
                        "name": "Request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/request.OrderBrc20PushReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/respond.Message"
                        }
                    }
                }
            }
        },
        "/brc20/orders": {
            "get": {
                "description": "Fetch orders",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "brc20"
                ],
                "summary": "Fetch orders",
                "parameters": [
                    {
                        "type": "string",
                        "description": "tick",
                        "name": "tick",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "sellerAddress",
                        "name": "sellerAddress",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "buyerAddress",
                        "name": "buyerAddress",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "orderState",
                        "name": "orderState",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "orderType",
                        "name": "orderType",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "limit",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "flag",
                        "name": "flag",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "sortKey",
                        "name": "sortKey",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "sortType",
                        "name": "sortType",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/respond.OrderResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "request.OrderBrc20PushReq": {
            "type": "object",
            "properties": {
                "address": {
                    "type": "string"
                },
                "orderState": {
                    "description": "1-create",
                    "type": "integer"
                },
                "orderType": {
                    "description": "1-sell,2-buy",
                    "type": "integer"
                },
                "psbtRaw": {
                    "type": "string"
                },
                "tick": {
                    "type": "string"
                }
            }
        },
        "respond.Brc20Item": {
            "type": "object",
            "properties": {
                "amount": {
                    "type": "integer"
                },
                "buyerAddress": {
                    "type": "string"
                },
                "coinAmount": {
                    "type": "integer"
                },
                "coinDecimalNum": {
                    "type": "integer"
                },
                "coinRatePrice": {
                    "type": "integer"
                },
                "decimalNum": {
                    "type": "integer"
                },
                "orderState": {
                    "description": "1-create,2-finish,3-cancel",
                    "type": "integer"
                },
                "orderType": {
                    "description": "1-sell,2-buy",
                    "type": "integer"
                },
                "psbtRaw": {
                    "type": "string"
                },
                "sellerAddress": {
                    "type": "string"
                },
                "tick": {
                    "type": "string"
                },
                "timestamp": {
                    "type": "integer"
                }
            }
        },
        "respond.Message": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "data": {},
                "message": {
                    "type": "string"
                },
                "processingTime": {
                    "type": "integer"
                }
            }
        },
        "respond.OrderResponse": {
            "type": "object",
            "properties": {
                "flag": {
                    "type": "integer"
                },
                "results": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/respond.Brc20Item"
                    }
                },
                "total": {
                    "type": "integer"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "",
	BasePath:    "/book",
	Schemes:     []string{"https"},
	Title:       "OrderBook API Service",
	Description: "OrderBook API Service",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
		"escape": func(v interface{}) string {
			// escape tabs
			str := strings.Replace(v.(string), "\t", "\\t", -1)
			// replace " with \", and if that results in \\", replace that with \\\"
			str = strings.Replace(str, "\"", "\\\"", -1)
			return strings.Replace(str, "\\\\\"", "\\\\\\\"", -1)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register("swagger", &s{})
}