package api

import (
	"dunno/internal/aws"
	"dunno/internal/log"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/go-chi/chi/v5"
)

type responseWriterDecorator struct {
	http.ResponseWriter
	StatusCode int
}

func (w *responseWriterDecorator) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func endpointDurationMiddleware(namespace, metric string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			decorator := responseWriterDecorator{
				ResponseWriter: w,
				StatusCode:     200,
			}
			before := time.Now()
			next.ServeHTTP(&decorator, r)
			duration := time.Now().Sub(before).Milliseconds()
			code := decorator.StatusCode
			pattern := strings.Join(chi.RouteContext(r.Context()).RoutePatterns, "")
			log.Logger.Infof("Pattern %s finished with code %d, duration: %d", pattern, code, duration)
			codeFamily := fmt.Sprintf("%cXX", []rune(strconv.Itoa(code))[0])
			_, err := aws.CloudWatchClient.PutMetricData(r.Context(), &cloudwatch.PutMetricDataInput{
				Namespace: awssdk.String(namespace),
				MetricData: []types.MetricDatum{
					{
						MetricName: awssdk.String(metric),
						Unit:       types.StandardUnitMilliseconds,
						Value:      awssdk.Float64(float64(duration)),
						Dimensions: []types.Dimension{
							{
								Name:  awssdk.String("pattern"),
								Value: awssdk.String(pattern),
							},
							{
								Name:  awssdk.String("codeFamily"),
								Value: awssdk.String(codeFamily),
							},
						},
					},
				},
			})
			if err != nil {
				log.Logger.Errorf("PutMetricData failed, error: %v", err)
				return
			}
		})
	}
}
