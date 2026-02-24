.PHONY: mailhog integration-test

mailhog:
	docker run --rm --name mailhog -p 1025:1025 -p 8025:8025 mailhog/mailhog

integration-test:
	@echo "Starting MailHog..."
	docker run -d --rm --name mailhog -p 1025:1025 -p 8025:8025 mailhog/mailhog
	@echo "Waiting for MailHog to be ready..."
	@sleep 2
	@echo "Running tests..."
	@go test -v -count=1 -tags=integration ./... ; \
	STATUS=$$? ; \
	echo "Stopping MailHog..." ; \
	docker stop mailhog ; \
	exit $$STATUS