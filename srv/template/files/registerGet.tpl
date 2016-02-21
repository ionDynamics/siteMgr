{{template "header" .}}
	<form class="form-register pp-login" method="post" action="/register">
		<h2 class="form-register-heading">Register</h2>
		<label for="login-name" class="sr-only">E-Mail</label>
		<input type="text" id="login-name" name="login-name" class="form-control" placeholder="E-Mail" required autofocus>
		<label for="login-pass" class="sr-only">Password</label>
		<input type="password" id="login-pass" name="login-pass" class="form-control" placeholder="Password" required>
		<label for="login-pass2" class="sr-only">Repeat Password</label>
		<input type="password" id="login-pass2" name="login-pass2" class="form-control" placeholder="Repeat Password" required>
		<button class="btn btn-lg btn-primary btn-block" type="submit">Register</button>
	</form>
{{template "footer" .}}