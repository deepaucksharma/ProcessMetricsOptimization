package adapter

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"

	"github.com/deepaucksharma/reservoir"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// OTelPDataAdapter converts between OpenTelemetry pdata types and reservoir domain models
type OTelPDataAdapter struct{}

// NewOTelPDataAdapter creates a new OpenTelemetry pdata adapter
func NewOTelPDataAdapter() *OTelPDataAdapter {
	return &OTelPDataAdapter{}
}

// ConvertSpan converts an OTEL span to a domain model span
func (a *OTelPDataAdapter) ConvertSpan(
	span interface{},
	resource interface{},
	scope interface{},
) reservoir.SpanData {
	// Type assertion
	pSpan, ok := span.(ptrace.Span)
	if !ok {
		return reservoir.SpanData{}
	}
	
	pResource, ok := resource.(pcommon.Resource)
	if !ok {
		return reservoir.SpanData{}
	}
	
	pScope, ok := scope.(pcommon.InstrumentationScope)
	if !ok {
		return reservoir.SpanData{}
	}
	// Extract span data
	traceID := pSpan.TraceID().String()
	spanID := pSpan.SpanID().String()
	
	var parentSpanID string
	if !pSpan.ParentSpanID().IsEmpty() {
		parentSpanID = pSpan.ParentSpanID().String()
	}
	
	// Extract attributes
	attrs := make(map[string]interface{})
	pSpan.Attributes().Range(func(k string, v pcommon.Value) bool {
		switch v.Type() {
		case pcommon.ValueTypeStr:
			attrs[k] = v.Str()
		case pcommon.ValueTypeInt:
			attrs[k] = v.Int()
		case pcommon.ValueTypeDouble:
			attrs[k] = v.Double()
		case pcommon.ValueTypeBool:
			attrs[k] = v.Bool()
		// Handle other types including slices and maps
		}
		return true
	})
	
	// Convert events
	events := make([]reservoir.Event, pSpan.Events().Len())
	for i := 0; i < pSpan.Events().Len(); i++ {
		srcEvent := pSpan.Events().At(i)
		eventAttrs := make(map[string]interface{})
		
		srcEvent.Attributes().Range(func(k string, v pcommon.Value) bool {
			switch v.Type() {
			case pcommon.ValueTypeStr:
				eventAttrs[k] = v.Str()
			case pcommon.ValueTypeInt:
				eventAttrs[k] = v.Int()
			case pcommon.ValueTypeDouble:
				eventAttrs[k] = v.Double()
			case pcommon.ValueTypeBool:
				eventAttrs[k] = v.Bool()
			// Handle other types
			}
			return true
		})
		
		event := reservoir.Event{
			Name:       srcEvent.Name(),
			Timestamp:  srcEvent.Timestamp().AsTime().UnixNano(),
			Attributes: eventAttrs,
		}
		events[i] = event
	}
	
	// Convert links
	links := make([]reservoir.Link, pSpan.Links().Len())
	for i := 0; i < pSpan.Links().Len(); i++ {
		srcLink := pSpan.Links().At(i)
		linkAttrs := make(map[string]interface{})
		
		srcLink.Attributes().Range(func(k string, v pcommon.Value) bool {
			switch v.Type() {
			case pcommon.ValueTypeStr:
				linkAttrs[k] = v.Str()
			case pcommon.ValueTypeInt:
				linkAttrs[k] = v.Int()
			case pcommon.ValueTypeDouble:
				linkAttrs[k] = v.Double()
			case pcommon.ValueTypeBool:
				linkAttrs[k] = v.Bool()
			// Handle other types
			}
			return true
		})
		
		link := reservoir.Link{
			TraceID:    srcLink.TraceID().String(),
			SpanID:     srcLink.SpanID().String(),
			Attributes: linkAttrs,
		}
		links[i] = link
	}
	
	// Create resource info
	resourceAttrs := make(map[string]interface{})
	pResource.Attributes().Range(func(k string, v pcommon.Value) bool {
		switch v.Type() {
		case pcommon.ValueTypeStr:
			resourceAttrs[k] = v.Str()
		case pcommon.ValueTypeInt:
			resourceAttrs[k] = v.Int()
		case pcommon.ValueTypeDouble:
			resourceAttrs[k] = v.Double()
		case pcommon.ValueTypeBool:
			resourceAttrs[k] = v.Bool()
		// Handle other types
		}
		return true
	})
	
	resourceInfo := reservoir.ResourceInfo{
		Attributes: resourceAttrs,
	}
	
	// Create scope info
	scopeInfo := reservoir.ScopeInfo{
		Name:    pScope.Name(),
		Version: pScope.Version(),
	}
	
	// Create the span data object
	return reservoir.SpanData{
		ID:           spanID,
		TraceID:      traceID,
		ParentID:     parentSpanID,
		Name:         pSpan.Name(),
		StartTime:    pSpan.StartTimestamp().AsTime().UnixNano(),
		EndTime:      pSpan.EndTimestamp().AsTime().UnixNano(),
		Attributes:   attrs,
		Events:       events,
		Links:        links,
		StatusCode:   int(pSpan.Status().Code()),
		StatusMsg:    pSpan.Status().Message(),
		ResourceInfo: resourceInfo,
		ScopeInfo:    scopeInfo,
	}
}

// ConvertToOTEL converts domain model spans back to OTEL format
func (a *OTelPDataAdapter) ConvertToOTEL(spans []reservoir.SpanData) interface{} {
	traces := ptrace.NewTraces()
	
	// Group spans by resource to correctly structure the OTEL data
	resourceMap := make(map[string]map[string][]reservoir.SpanData)
	
	for _, span := range spans {
		// Create a key for the resource based on its attributes
		resourceKey := ""
		for k, v := range span.ResourceInfo.Attributes {
			resourceKey += k + ":" + fmt.Sprintf("%v", v) + ","
		}
		
		// Create a key for the scope
		scopeKey := span.ScopeInfo.Name + ":" + span.ScopeInfo.Version
		
		// Initialize maps if needed
		if _, ok := resourceMap[resourceKey]; !ok {
			resourceMap[resourceKey] = make(map[string][]reservoir.SpanData)
		}
		
		// Add the span to the appropriate resource and scope
		resourceMap[resourceKey][scopeKey] = append(resourceMap[resourceKey][scopeKey], span)
	}
	
	// Create OTEL structure
	for _, scopeMap := range resourceMap {
		// Create a new ResourceSpans
		rs := traces.ResourceSpans().AppendEmpty()
		
		// For each scope in this resource
		for scopeKey, scopeSpans := range scopeMap {
			// Create a new ScopeSpans
			ss := rs.ScopeSpans().AppendEmpty()
			scope := ss.Scope()
			
			// Set scope info
			// Parse scopeKey to get name and version
			scopeParts := strings.Split(scopeKey, ":")
			if len(scopeParts) == 2 {
				scope.SetName(scopeParts[0])
				scope.SetVersion(scopeParts[1])
			}
			
			// Add all spans for this scope
			for _, span := range scopeSpans {
				otelSpan := ss.Spans().AppendEmpty()
				
				// Set span data
				if traceID, err := hex.DecodeString(span.TraceID); err == nil && len(traceID) == 16 {
					var otelTraceID pcommon.TraceID
					copy(otelTraceID[:], traceID)
					otelSpan.SetTraceID(otelTraceID)
				}
				
				if spanID, err := hex.DecodeString(span.ID); err == nil && len(spanID) == 8 {
					var otelSpanID pcommon.SpanID
					copy(otelSpanID[:], spanID)
					otelSpan.SetSpanID(otelSpanID)
				}
				
				otelSpan.SetName(span.Name)
				otelSpan.SetStartTimestamp(pcommon.Timestamp(span.StartTime))
				otelSpan.SetEndTimestamp(pcommon.Timestamp(span.EndTime))
			}
		}
	}
	
	return traces
}

// ConvertTraces converts a complete OTEL traces object to domain model spans
func (a *OTelPDataAdapter) ConvertTraces(traces interface{}) []reservoir.SpanData {
	var result []reservoir.SpanData
	
	otelTraces, ok := traces.(ptrace.Traces)
	if !ok {
		return result
	}
	
	// Process each resource spans
	rss := otelTraces.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		resource := rs.Resource()
		
		// Process each instrumentation scope spans
		ilss := rs.ScopeSpans()
		for j := 0; j < ilss.Len(); j++ {
			ils := ilss.At(j)
			scope := ils.Scope()
			
			// Process each span
			spans := ils.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				result = append(result, a.ConvertSpan(span, resource, scope))
			}
		}
	}
	
	return result
}

// Helper function to set attribute values in OTEL maps
func setAttributeValue(attrMap pcommon.Map, key string, value interface{}) {
	switch v := value.(type) {
	case string:
		attrMap.PutStr(key, v)
	case int, int8, int16, int32, int64:
		attrMap.PutInt(key, reflect.ValueOf(v).Int())
	case uint, uint8, uint16, uint32, uint64:
		attrMap.PutInt(key, int64(reflect.ValueOf(v).Uint()))
	case float32, float64:
		attrMap.PutDouble(key, reflect.ValueOf(v).Float())
	case bool:
		attrMap.PutBool(key, v)
	default:
		// Attempt to convert to string
		attrMap.PutStr(key, fmt.Sprintf("%v", v))
	}
}