package customfilters

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/zalando/skipper/filters/filtertest"
)

func TestFilterSpec(t *testing.T) {
	expected := "sed"

	spec := NewSed()

	if spec.Name() != expected {
		t.Error("Expected " + expected + ", got " + spec.Name())
	}
}

func TestRegularExpressionSubstitution(t *testing.T) {

	response := `<?xml version='1.0' encoding='UTF-8'?>
        <wsdl:definitions xmlns:xsd="http://www.w3.org/2001/XMLSchema"
            xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/"
            xmlns:tns="http://reading.webservice.price.zalando.de/"
            xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
            xmlns:ns2="http://schemas.xmlsoap.org/soap/http"
            xmlns:ns1="http://price.zalando.de/read/csvExport"
            name="CSVPriceReadWebServiceImplService"
            targetNamespace="http://reading.webservice.price.zalando.de/">
            <wsdl:import location="http://restsn08.zalando:37007/ws/csvPriceReadWebService?wsdl=CSVPriceReadWebService.wsdl"
                namespace="http://price.zalando.de/read/csvExport">
            </wsdl:import>
            <wsdl:binding name="CSVPriceReadWebServiceImplServiceSoapBinding"
                type="ns1:CSVPriceReadWebService">
                <soap:binding style="document"
                    transport="http://schemas.xmlsoap.org/soap/http"/>
                <wsdl:operation
                    name="getWholeCurrentPromotionalBlacklistZipFile">
                    <soap:operation soapAction="" style="document"/>
                        <wsdl:input
                            name="getWholeCurrentPromotionalBlacklistZipFile">
                            <soap:body use="literal"/>
                        </wsdl:input>
                        <wsdl:output
                            name="getWholeCurrentPromotionalBlacklistZipFileResponse">
                            <soap:body use="literal"/>
                        </wsdl:output>
                </wsdl:operation>
            </wsdl:binding>
            <wsdl:service name="CSVPriceReadWebServiceImplService">
                <wsdl:port 
                    binding="tns:CSVPriceReadWebServiceImplServiceSoapBinding"
                    name="CSVPriceReadWebServiceImplPort">
                    <soap:address
                        location="http://restsn08.zalando:37077/ws/csvPriceReadWebService"/>
                </wsdl:port>
            </wsdl:service>
        </wsdl:definitions>`
	expected := `<?xml version='1.0' encoding='UTF-8'?>
        <wsdl:definitions xmlns:xsd="http://www.w3.org/2001/XMLSchema"
            xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/"
            xmlns:tns="http://reading.webservice.price.zalando.de/"
            xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
            xmlns:ns2="http://schemas.xmlsoap.org/soap/http"
            xmlns:ns1="http://price.zalando.de/read/csvExport"
            name="CSVPriceReadWebServiceImplService"
            targetNamespace="http://reading.webservice.price.zalando.de/">
            <wsdl:import location="https://price-service.tm.zalando.com:37077/ws/csvPriceReadWebService?wsdl=CSVPriceReadWebService.wsdl"
                namespace="http://price.zalando.de/read/csvExport">
            </wsdl:import>
            <wsdl:binding name="CSVPriceReadWebServiceImplServiceSoapBinding"
                type="ns1:CSVPriceReadWebService">
                <soap:binding style="document"
                    transport="http://schemas.xmlsoap.org/soap/http"/>
                <wsdl:operation
                    name="getWholeCurrentPromotionalBlacklistZipFile">
                    <soap:operation soapAction="" style="document"/>
                        <wsdl:input
                            name="getWholeCurrentPromotionalBlacklistZipFile">
                            <soap:body use="literal"/>
                        </wsdl:input>
                        <wsdl:output
                            name="getWholeCurrentPromotionalBlacklistZipFileResponse">
                            <soap:body use="literal"/>
                        </wsdl:output>
                </wsdl:operation>
            </wsdl:binding>
            <wsdl:service name="CSVPriceReadWebServiceImplService">
                <wsdl:port 
                    binding="tns:CSVPriceReadWebServiceImplServiceSoapBinding"
                    name="CSVPriceReadWebServiceImplPort">
                    <soap:address
                        location="https://price-service.tm.zalando.com:37077/ws/csvPriceReadWebService"/>
                </wsdl:port>
            </wsdl:service>
        </wsdl:definitions>`
	resp := &http.Response{Body: ioutil.NopCloser(strings.NewReader(response)),
		ContentLength: int64(len(response))}

	sp := NewSed()
	conf := []interface{}{
		"location=\"https?://[^/]+/ws/",
		"location=\"https://price-service.tm.zalando.com:37077/ws/"}
	f, err := sp.CreateFilter(conf)
	if err != nil {
		t.Error(err)
	}

	ctx := &filtertest.Context{FResponse: resp}
	f.Response(ctx)

	body, err := ioutil.ReadAll(ctx.Response().Body)
	if err != nil {
		t.Error(err)
	}

	if resp.ContentLength != int64(len(expected)) {
		t.Error("Content length does not match.")
	}
	if string(body) != expected {
		t.Error("Expected \"" + expected + "\", got \"" + string(body) + "\"")
	}
}
