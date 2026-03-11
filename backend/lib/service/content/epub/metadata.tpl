<div
    style="
        font-size: 0.85em;
        color: #666;
        margin-bottom: 2em;
        padding: 1em;
        border-left: 3px solid #ccc;
        background-color: #f9f9f9;
    "
>
    {{- if .Title}}
    <p><strong>Title: {{.Title}}</strong></p>
    {{- end}}{{- if .Author}}
    <p><strong>Author: {{.Author}}</strong></p>
    {{- end}}{{- if (sourceInfo .)}}
    <p><strong>Source: {{sourceInfo .}}</strong></p>
    {{- end}}{{- if gt .ReadingTimeMinutes 0}}
    <p><strong>Reading time: {{.ReadingTimeMinutes}} min</strong></p>
    {{- end}}{{- if and .PublishedAt (not .PublishedAt.IsZero)}}
    <p>
        <strong
            >Published: {{.PublishedAt.Format
            "2006-01-02T15:04:05Z07:00"}}</strong
        >
    </p>
    {{- end}}{{- if not .CreatedAt.IsZero}}
    <p>
        <strong
            >Added: {{.CreatedAt.Format "2006-01-02T15:04:05Z07:00"}}</strong
        >
    </p>
    {{- end}}
</div>
