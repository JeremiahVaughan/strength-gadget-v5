package main

import (
	"fmt"
	"html/template"
)

func registerTemplates() error {
	formTemplates := []string{
		"templates/components/button.html",
		"templates/components/form-header.html",
		"templates/components/text-input.html",
	}
	landingPageTemplatesParsed, err := template.ParseFS(
		templatesFiles,
		"templates/landing-page.html",
		"templates/base.html",
		"templates/components/button.html",
	)
	if err != nil {
		return fmt.Errorf("error, when attempting to parse html landingPageTemplates. Error: %v", err)
	}
	templateMap["landing-page.html"] = landingPageTemplatesParsed

	signUpFormTemplates := append(
		formTemplates,
		"templates/components/sign-up-form.html",
	)
	var signUpFormTemplatesParsed *template.Template
	signUpFormTemplatesParsed, err = template.ParseFS(
		templatesFiles,
		signUpFormTemplates...,
	)
	if err != nil {
		return fmt.Errorf("error, when attempting to parse html signUpFormTemplates. Error: %v", err)
	}
	templateMap["sign-up-form.html"] = signUpFormTemplatesParsed

	var signUpPageTemplatesParsed *template.Template
	signUpPageTemplatesParsed, err = template.ParseFS(
		templatesFiles,
		append(
			signUpFormTemplates,
			"templates/sign-up-page.html",
			"templates/base.html",
		)...,
	)
	if err != nil {
		return fmt.Errorf("error, when attempting to parse html signUpPageTemplates. Error: %v", err)
	}
	templateMap["sign-up-page.html"] = signUpPageTemplatesParsed

	verificationFormTemplates := append(
		formTemplates,
		"templates/components/verification-form.html",
	)
	var verificationFormTemplatesParsed *template.Template
	verificationFormTemplatesParsed, err = template.ParseFS(
		templatesFiles,
		verificationFormTemplates...,
	)
	if err != nil {
		return fmt.Errorf("error, when attempting to parse html verificationFormTemplates. Error: %v", err)
	}
	templateMap["verification-form.html"] = verificationFormTemplatesParsed

	var verificationPageTemplatesParsed *template.Template
	verificationPageTemplatesParsed, err = template.ParseFS(
		templatesFiles,
		append(
			verificationFormTemplates,
			"templates/verification-page.html",
			"templates/base.html",
		)...,
	)
	if err != nil {
		return fmt.Errorf("error, when attempting to parse html verificationPageTemplatesParsed. Error: %v", err)
	}
	templateMap["verification-page.html"] = verificationPageTemplatesParsed

	loginFormTemplates := append(
		formTemplates,
		"templates/components/login-form.html",
	)
	var loginFormTemplatesParsed *template.Template
	loginFormTemplatesParsed, err = template.ParseFS(
		templatesFiles,
		loginFormTemplates...,
	)
	if err != nil {
		return fmt.Errorf("error, when attempting to parse html loginFormTemplates. Error: %v", err)
	}
	templateMap["login-form.html"] = loginFormTemplatesParsed

	var loginPageTemplatesParsed *template.Template
	loginPageTemplatesParsed, err = template.ParseFS(
		templatesFiles,
		append(
			loginFormTemplates,
			"templates/login-page.html",
			"templates/base.html",
		)...,
	)
	if err != nil {
		return fmt.Errorf("error, when attempting to parse html loginPageTemplatesParsed. Error: %v", err)
	}
	templateMap["login-page.html"] = loginPageTemplatesParsed

	forgotPasswordEmailFormTemplates := append(
		formTemplates,
		"templates/components/forgot-password-email-form.html",
	)
	var forgotPasswordEmailFormParsed *template.Template
	forgotPasswordEmailFormParsed, err = template.ParseFS(
		templatesFiles,
		forgotPasswordEmailFormTemplates...,
	)
	if err != nil {
		return fmt.Errorf("error, when attempting to parse html forgotPasswordFormEmailParsed. Error: %v", err)
	}
	templateMap["forgot-password-email-form.html"] = forgotPasswordEmailFormParsed

	var forgotPasswordEmailPageParsed *template.Template
	forgotPasswordEmailPageParsed, err = template.ParseFS(
		templatesFiles,
		append(
			forgotPasswordEmailFormTemplates,
			"templates/forgot-password-email-page.html",
			"templates/base.html",
		)...,
	)
	if err != nil {
		return fmt.Errorf("error, when atempting to parse html forgotPasswordEmailPageParsed. Error: %v", err)
	}
	templateMap["forgot-password-email-page.html"] = forgotPasswordEmailPageParsed

	forgotPasswordResetCodeFormTemplates := append(
		formTemplates,
		"templates/components/forgot-password-reset-code-form.html",
	)
	var forgotPasswordResetCodeFormParsed *template.Template
	forgotPasswordResetCodeFormParsed, err = template.ParseFS(
		templatesFiles,
		forgotPasswordResetCodeFormTemplates...,
	)
	if err != nil {
		return fmt.Errorf("error, when attempting to parse html forgotPasswordResetCodeFormParsed. Error: %v", err)
	}
	templateMap["forgot-password-reset-code-form.html"] = forgotPasswordResetCodeFormParsed

	var forgotPasswordResetCodePageParsed *template.Template
	forgotPasswordResetCodePageParsed, err = template.ParseFS(
		templatesFiles,
		append(
			forgotPasswordResetCodeFormTemplates,
			"templates/forgot-password-reset-code-page.html",
			"templates/base.html",
		)...,
	)
	if err != nil {
		return fmt.Errorf("error, when atempting to parse html forgotPasswordPageResetCodeParsed. Error: %v", err)
	}
	templateMap["forgot-password-reset-code-page.html"] = forgotPasswordResetCodePageParsed

	forgotPasswordNewPasswordFormTemplates := append(
		formTemplates,
		"templates/components/forgot-password-new-password-form.html",
	)
	var forgotPasswordNewPasswordFormParsed *template.Template
	forgotPasswordNewPasswordFormParsed, err = template.ParseFS(
		templatesFiles,
		forgotPasswordNewPasswordFormTemplates...,
	)
	if err != nil {
		return fmt.Errorf("error, when attempting to parse html forgotPasswordFormNewPasswordParsed. Error: %v", err)
	}
	templateMap["forgot-password-new-password-form.html"] = forgotPasswordNewPasswordFormParsed

	var forgotPasswordPageNewPasswordParsed *template.Template
	forgotPasswordPageNewPasswordParsed, err = template.ParseFS(
		templatesFiles,
		append(
			forgotPasswordNewPasswordFormTemplates,
			"templates/forgot-password-new-password-page.html",
			"templates/base.html",
		)...,
	)
	if err != nil {
		return fmt.Errorf("error, when attempting to parse html forgotPasswordPageNewPasswordParsed. Error: %v", err)
	}
	templateMap["forgot-password-new-password-page.html"] = forgotPasswordPageNewPasswordParsed

	var exercisePageParsed *template.Template
	exercisePageParsed, err = template.ParseFS(
		templatesFiles,
		"templates/base.html",
		"templates/exercise-page.html",
		"templates/components/button.html",
		"templates/components/tool-bar.html",
	)
	if err != nil {
		return fmt.Errorf("error, when attempting to parse html exercisePageParsed. Error: %v", err)
	}
	templateMap["exercise-page.html"] = exercisePageParsed

	var workoutCompletedParsed *template.Template
	workoutCompletedParsed, err = template.ParseFS(
		templatesFiles,
		"templates/base.html",
		"templates/workout-completed-page.html",
		"templates/components/form-header.html",
		"templates/components/button.html",
		"templates/components/tool-bar.html",
	)
	if err != nil {
		return fmt.Errorf("error, when attempting to parse html workoutCompletedParsed. Error: %v", err)
	}
	templateMap["workout-completed-page.html"] = workoutCompletedParsed

	return nil
}
