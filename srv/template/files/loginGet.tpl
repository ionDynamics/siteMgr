{{template "header" .}}
	<form class="form-signin sm-login" method="post" action="/login" autocomplete="on">

		<h2 class="form-signin-heading">Log in</h2>

		<div class="form-group">
			<label for="login-name" class="sr-only">E-Mail</label>
			<input type="text" id="login-name" name="login-name" class="form-control" placeholder="E-Mail" required autofocus>
		</div>

		<div class="form-group">
			<label for="login-pass" class="sr-only">Password</label>
			<input type="password" id="login-pass" name="login-pass" class="form-control" placeholder="Password" required>
		</div>

		<button class="btn btn-lg btn-primary btn-block" type="submit">Log in</button>
	</form>
{{template "footer" .}}