# --- Build Stage ---
FROM golang:1.24 AS build

ENV CODE_DIR=/go/src/
WORKDIR ${CODE_DIR}

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

ARG LDFLAGS

RUN CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o /opt/calendar/calendar ./cmd/calendar
RUN CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o /opt/calendar/calendar-scheduler ./cmd/calendar_scheduler
RUN CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o /opt/calendar/calendar-sender ./cmd/calendar_sender

FROM scratch

LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="calendar"
LABEL MAINTAINERS="mi.volkov@bi.zone"

COPY --from=build /opt/calendar/calendar /opt/calendar/calendar
COPY --from=build /opt/calendar/calendar-scheduler /opt/calendar/calendar-scheduler
COPY --from=build /opt/calendar/calendar-sender /opt/calendar/calendar-sender

COPY ./configs /etc/calendar/

