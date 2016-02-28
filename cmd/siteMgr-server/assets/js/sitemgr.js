$(document).ready(function (){
	$(".delBtn").click(function(e) {
		e.preventDefault()
		$(this).parent().children(".confirmBtn").toggleClass("hidden")
	})
})
