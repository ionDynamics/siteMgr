$(document).ready(function (){
	$("#delBtn").click(function(e) {
		e.preventDefault()
		$("#confirmBtn").toggleClass("hidden")
	})
})
