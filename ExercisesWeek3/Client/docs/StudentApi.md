# {{classname}}

All URIs are relative to *https://itu.dk/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddStudent**](StudentApi.md#AddStudent) | **Post** /student | Add a new student
[**DeleteStudent**](StudentApi.md#DeleteStudent) | **Delete** /student/{studentId} | Deletes a student
[**GetStudentById**](StudentApi.md#GetStudentById) | **Get** /student/{studentId} | Find student by ID
[**UpdateStudent**](StudentApi.md#UpdateStudent) | **Put** /student | Update a student

# **AddStudent**
> Student AddStudent(ctx, body, studentId, name, courses)
Add a new student

Add a new student

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Student**](Student.md)| Add a new student | 
  **studentId** | **int64**|  | 
  **name** | **string**|  | 
  **courses** | [**[]Course**](Course.md)|  | 

### Return type

[**Student**](student.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json, application/xml, application/x-www-form-urlencoded
 - **Accept**: application/json, application/xml

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteStudent**
> DeleteStudent(ctx, studentId)
Deletes a student

delete a student

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **studentId** | **int64**| Student id to delete | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetStudentById**
> Student GetStudentById(ctx, studentId)
Find student by ID

Returns a single student

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **studentId** | **int64**| ID of student to return | 

### Return type

[**Student**](student.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/xml

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateStudent**
> Student UpdateStudent(ctx, body, studentId, name, courses)
Update a student

Update an existing student by Id

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Student**](Student.md)| Update an existent student | 
  **studentId** | **int64**|  | 
  **name** | **string**|  | 
  **courses** | [**[]Course**](Course.md)|  | 

### Return type

[**Student**](student.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json, application/xml, application/x-www-form-urlencoded
 - **Accept**: application/json, application/xml

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

