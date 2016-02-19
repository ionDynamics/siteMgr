{{template "header" .}}
	<form class="form-register pp-login" method="post" action="/register">
		<h2 class="form-register-heading">Registrieren</h2>
		<label for="login-name" class="sr-only">E-Mail</label>
		<input type="text" id="login-name" name="login-name" class="form-control" placeholder="E-Mail" required autofocus>
		<label for="login-pass" class="sr-only">Passwort</label>
		<input type="password" id="login-pass" name="login-pass" class="form-control" placeholder="Passwort" required>
		<label for="login-pass2" class="sr-only">Passwort wiederholen</label>
		<input type="password" id="login-pass2" name="login-pass2" class="form-control" placeholder="Passwort wiederholen" required>
		<button class="btn btn-lg btn-primary btn-block" type="submit">Registrieren</button>
	</form>
{{template "footer" .}}