# {{classname}}

All URIs are relative to *https://itu.dk/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CourseCourseIdDelete**](DefaultApi.md#CourseCourseIdDelete) | **Delete** /course/{courseId} | 
[**CourseCourseIdGet**](DefaultApi.md#CourseCourseIdGet) | **Get** /course/{courseId} | 
[**CoursePost**](DefaultApi.md#CoursePost) | **Post** /course | 

# **CourseCourseIdDelete**
> CourseCourseIdDelete(ctx, courseId)


Delete a course

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **courseId** | **int64**| courseId to delete | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **CourseCourseIdGet**
> CourseCourseIdGet(ctx, courseId)


Get info about a course

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **courseId** | **int64**| courseId to get | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **CoursePost**
> CoursePost(ctx, courseId)


create a new course

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **courseId** | **int64**| courseId to create | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

