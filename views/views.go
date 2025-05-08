package views


import (
    "log"
    "strengthgadget.com/m/v2/ui_util"
)

type Views struct {

}

func New() Views {
    templates := []ui_util.HtmlTemplate{
        {
            Name: "landing-page",
            FileOverrides: []string{
                "landing-page.html",
            },
        },
        {
            Name: "exercise-page",
            FileOverrides: []string{
                "exercise-page.html",
            },
        },
        {
            Name: "workout-completed-page",
            FileOverrides: []string{
                "workout-completed-page.html",
            },
        },
    }
    tl, err = ui_util.NewTemplateLoader(
        "templates/base",
        "templates/overrides",
        templates,
        localMode,
    )
    if err != nil {
        log.Fatalf("error, when ui_util.NewTemplateLoader() for main(). Error: %v", err)
    }
}
