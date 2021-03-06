package information

import (
	"net/http"
	"strconv"

	"github.com/asepnur/meiko_information/src/webserver/template"

	"github.com/asepnur/meiko_information/src/module/course"
	cs "github.com/asepnur/meiko_information/src/module/course"
	inf "github.com/asepnur/meiko_information/src/module/information"
	rg "github.com/asepnur/meiko_information/src/module/rolegroup"
	"github.com/asepnur/meiko_information/src/util/auth"
	"github.com/julienschmidt/httprouter"
)

// GetHandler ...
func GetHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	sess := r.Context().Value("User").(*auth.User)

	params := getParams{
		total: r.FormValue("ttl"),
		page:  r.FormValue("pg"),
	}

	args, err := params.validate()
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError("Invalid request"))
		return
	}

	courses, err := cs.SelectEnrolledSchedule(sess.ID)
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusInternalServerError))
		return
	}
	scheduleID := []int64{}
	for _, val := range courses {
		scheduleID = append(scheduleID, val.Schedule.ID)
	}
	offset := (args.page - 1) * args.total
	rows, err := inf.CountInformation(scheduleID, args.total, offset)
	totalPage := rows / 10
	rest := rows % 10
	if rest > 0 {
		totalPage++
	}
	meta := meta{
		TotalPage: totalPage,
	}
	links := links{
		Self: args.page,
		Next: (args.page + 1),
		Prev: (args.page - 1),
	}
	informations, err := inf.SelectByPage(scheduleID, args.total, offset)
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError(err.Error()))
		return
	}

	informationsID := []string{}
	for _, val := range informations {
		informationsID = append(informationsID, strconv.FormatInt(val.ID, 10))
	}

	data := []dataList{}
	for _, val := range informations {
		courseName := "general"
		if val.ScheduleID.Int64 != 0 {
			for _, c := range courses {
				if c.Schedule.ID == val.ScheduleID.Int64 {
					courseName = c.Course.Name
				}
			}
		}
		data = append(data, dataList{
			ID:          val.ID,
			Title:       val.Title,
			Description: val.Description.String,
			UpdatedDate: val.CreatedAt.Format("Monday, 2 January 2006"),
			CourseName:  courseName,
		})
	}
	res := respListInformation{
		Data:  data,
		Meta:  meta,
		Links: links,
	}
	template.RenderJSONResponse(w, new(template.Response).
		SetCode(http.StatusOK).
		SetData(res))
	return
}

// CreateHandler func ...
func CreateHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sess := r.Context().Value("User").(*auth.User)
	if !sess.IsHasRoles(rg.ModuleInformation, rg.RoleCreate, rg.RoleXCreate) {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusForbidden).
			AddError("You don't have privilege"))
		return
	}
	params := createParams{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		ScheduleID:  r.FormValue("schedule_id"),
	}
	args, err := params.validate()
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError(err.Error()))
		return
	}
	if args.ScheduleID != 0 {
		if !cs.IsAssistant(sess.ID, args.ScheduleID) {
			template.RenderJSONResponse(w, new(template.Response).
				SetCode(http.StatusBadRequest).
				AddError("You do not have privilage for this course!"))
			return
		}
	}
	err = inf.Insert(args.Title, args.Description, args.ScheduleID)
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError(err.Error()))
		return
	}

	template.RenderJSONResponse(w, new(template.Response).
		SetCode(http.StatusOK).
		SetMessage("Information created successfully"))
	return

}

// UpdateHandler func ...
func UpdateHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sess := r.Context().Value("User").(*auth.User)
	if !sess.IsHasRoles(rg.ModuleInformation, rg.RoleUpdate, rg.RoleXUpdate) {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusForbidden).
			AddError("You don't have privilege"))
		return
	}
	params := updateParams{
		ID:          ps.ByName("id"),
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		ScheduleID:  r.FormValue("schedule_id"),
	}

	args, err := params.validate()
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError(err.Error()))
		return
	}
	if args.ScheduleID != 0 {
		if !cs.IsAssistant(sess.ID, args.ScheduleID) {
			template.RenderJSONResponse(w, new(template.Response).
				SetCode(http.StatusBadRequest).
				AddError("Schedule ID forbidden!"))
			return
		}
	}

	if !inf.IsInformationIDExist(args.ID) {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError("Information ID does not exist"))
		return
	}

	if args.ScheduleID != 0 {
		// no validation
		// if !cs.IsExistScheduleID(args.ScheduleID) {
		// 	template.RenderJSONResponse(w, new(template.Response).
		// 		SetCode(http.StatusBadRequest).
		// 		AddError("Schedule ID does not exist"))
		// 	return
		// }
	}
	err = inf.Update(args.Title, args.Description, args.ScheduleID, args.ID)
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusInternalServerError))
		return
	}
	template.RenderJSONResponse(w, new(template.Response).
		SetCode(http.StatusOK).
		SetMessage("Update information succesfully"))
	return
}

// AvailableCourseInformation func
func AvailableCourseInformation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sess := r.Context().Value("User").(*auth.User)
	if !sess.IsHasRoles(rg.ModuleInformation, rg.RoleCreate, rg.RoleXCreate) {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusForbidden).
			AddError("You don't have privilege"))
		return
	}

	courses, err := cs.SelectTeachedSchedule(sess.ID)
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusInternalServerError))
		return
	}

	res := []respAvailableCourse{}
	for _, val := range courses {
		res = append(res, respAvailableCourse{
			ScheduleID: val.Schedule.ID,
			CourseName: val.Course.Name,
		})
	}
	template.RenderJSONResponse(w, new(template.Response).
		SetCode(http.StatusOK).
		SetData(res))
	return
}

// GetDetailByAdminHandler func ...
func GetDetailByAdminHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sess := r.Context().Value("User").(*auth.User)
	if !sess.IsHasRoles(rg.ModuleInformation, rg.RoleRead, rg.RoleXRead) {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusForbidden).
			AddError("You don't have privilege"))
		return
	}
	params := detailInfromationParams{
		ID: ps.ByName("id"),
	}
	args, err := params.validate()
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError(err.Error()))
		return
	}

	res, err := inf.GetByID(args.ID)
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError(err.Error()))
		return
	}

	id := res.ScheduleID.Int64
	desc := "-"
	course := cs.GetOne{}
	if id != 0 {
		course, err = cs.Get(sess.ID, id, true)
		if err != nil {
			template.RenderJSONResponse(w, new(template.Response).
				SetCode(http.StatusInternalServerError))
			return
		}
		if !course.Involved {
			template.RenderJSONResponse(w, new(template.Response).
				SetCode(http.StatusBadRequest).
				AddError("Not Authorized"))
			return
		}
	}

	if res.Description.Valid {
		desc = res.Description.String
	}
	response := respDetailInformation{
		ID:          res.ID,
		Title:       res.Title,
		Description: desc,
		CreatedDate: res.CreatedAt.Format("Monday, 2 January 2006 15:04:05"),
		UpdatedDate: res.UpdatedAt.Format("Monday, 2 January 2006 15:04:05"),
		Date:        res.UpdatedAt,
		ScheduleID:  id,
		CourseName:  course.Course.Course.Name,
	}
	template.RenderJSONResponse(w, new(template.Response).
		SetCode(http.StatusOK).
		SetData(response))
	return

}

// ReadHandler func ...
func ReadHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sess := r.Context().Value("User").(*auth.User)

	params := readListParams{
		total: r.FormValue("ttl"),
		page:  r.FormValue("pg"),
	}

	args, err := params.validate()
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError("Invalid request"))
		return
	}

	courses, err := cs.SelectTeachedSchedule(sess.ID)
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusInternalServerError))
		return
	}

	scheduleID := []int64{}
	for _, val := range courses {
		scheduleID = append(scheduleID, val.Schedule.ID)
	}

	offset := (args.page - 1) * args.total
	rows, err := inf.CountInformation(scheduleID, args.total, offset)
	totalPage := rows / 10
	rest := rows % 10
	if rest > 0 {
		totalPage++
	}
	meta := meta{
		TotalPage: totalPage,
	}
	links := links{
		Self: args.page,
		Next: (args.page + 1),
		Prev: (args.page - 1),
	}
	result, err := inf.SelectByPage(scheduleID, args.total, offset)
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError(err.Error()))
		return
	}
	data := []dataList{}
	for _, value := range result {
		courseName := "general"
		if value.ScheduleID.Valid {
			for _, val := range courses {
				if value.ScheduleID.Int64 == val.Schedule.ID {
					courseName = val.Course.Name
				}
			}
		}
		data = append(data, dataList{
			ID:          value.ID,
			Title:       value.Title,
			CreatedDate: value.CreatedAt.Format("Monday, 2 January 2006 15:04:05"),
			UpdatedDate: value.UpdatedAt.Format("Monday, 2 January 2006 15:04:05"),
			CourseName:  courseName,
		})
	}
	res := respListInformation{
		Data:  data,
		Meta:  meta,
		Links: links,
	}
	template.RenderJSONResponse(w, new(template.Response).
		SetCode(http.StatusOK).
		SetData(res))
	return
}

// DeleteHandler func ...
func DeleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sess := r.Context().Value("User").(*auth.User)
	if !sess.IsHasRoles(rg.ModuleInformation, rg.RoleDelete, rg.RoleXDelete) {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusForbidden).
			AddError("You don't have privilege"))
		return
	}
	params := deleteParams{
		ID: ps.ByName("id"),
	}
	args, err := params.validate()
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError(err.Error()))
		return
	}

	if !inf.IsInformationIDExist(args.ID) {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError("Information ID does not exist"))
		return
	}
	scheduleID := inf.GetScheduleIDByID(args.ID)
	if scheduleID != 0 && !cs.IsAssistant(sess.ID, scheduleID) {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError("You do not privilage for this informations"))
		return
	}

	err = inf.Delete(args.ID)
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusInternalServerError))
		return
	}
	template.RenderJSONResponse(w, new(template.Response).
		SetCode(http.StatusOK).
		SetMessage("Delete information successfully"))
	return

}

// GetDetailHandler func ..
func GetDetailHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	sess := r.Context().Value("User").(*auth.User)
	params := getDetailParams{
		id: ps.ByName("id"),
	}

	args, err := params.validate()
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).AddError(err.Error()))
		return
	}

	scheduleID := inf.GetScheduleIDByID(args.id)
	if scheduleID != 0 && !course.IsEnrolled(sess.ID, scheduleID) {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError("you do not have permission to this informations"))
		return
	}

	information, err := inf.GetByID(args.id)
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError("Information does not exist"))
		return
	}

	desc := "-"
	if information.Description.Valid {
		desc = information.Description.String
	}

	resp := getDetailResponse{
		ID:          information.ID,
		Title:       information.Title,
		Description: desc,
		Date:        information.CreatedAt.Format("Monday, 2 January 2006"),
	}

	template.RenderJSONResponse(w, new(template.Response).
		SetCode(http.StatusOK).
		SetData(resp))
	return
}
