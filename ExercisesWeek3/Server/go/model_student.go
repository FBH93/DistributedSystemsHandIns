/*
 * ITU REST service 1.0
 *
 * Hello, World!
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type Student struct {

	StudentId int64 `json:"studentId,omitempty"`

	Name string `json:"name,omitempty"`

	Courses []Course `json:"courses,omitempty"`
}
