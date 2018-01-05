package webserver

import (
	"github.com/asepnur/meiko_information/src/util/auth"
	"github.com/asepnur/meiko_information/src/webserver/handler/information"
	"github.com/julienschmidt/httprouter"
)

// Load returns all routing of this server
func loadRouter(r *httprouter.Router) {

	// ======================= Information Handler ======================
	// Admin section
	r.POST("/api/admin/v1/information", auth.MustAuthorize(information.CreateHandler))                  // create infomation by admin
	r.PATCH("/api/admin/v1/information/:id", auth.MustAuthorize(information.UpdateHandler))             // update infomation by admin
	r.DELETE("/api/admin/v1/information/:id", auth.MustAuthorize(information.DeleteHandler))            // delete information by admin
	r.GET("/api/admin/v1/information", auth.MustAuthorize(information.ReadHandler))                     // read list information by admin
	r.GET("/api/admin/v1/information/:id", auth.MustAuthorize(information.GetDetailByAdminHandler))     // read detail information by admin
	r.GET("/api/admin/v1/available-course", auth.MustAuthorize(information.AvailableCourseInformation)) // read detail information by admin
	// User section
	r.GET("/api/v1/information", auth.MustAuthorize(information.GetHandler))           // list informations
	r.GET("/api/v1/information/:id", auth.MustAuthorize(information.GetDetailHandler)) // detail information
	// ===================== End Information Handler ====================

}
