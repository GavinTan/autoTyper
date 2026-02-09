app_name = autoTyper
app_version = 2.0.3
app_id = com.tw.autoTyper


build:
	go build


darwin:
	fyne-cross darwin -arch amd64 -app-id $(app_id) -app-version $(app_version) -icon ./icon.png -name $(app_name) -macosx-version-min 12

darwin_arm:
	fyne-cross darwin -arch arm64 -app-id $(app_id) -app-version $(app_version) -icon ./icon.png -name $(app_name) -macosx-version-min 12

create_dmg:
	mkdir tmp && mv ./release/$(app_name).app ./tmp/
	chmod +x ./tmp/$(app_name).app/Contents/MacOS/$(app_name)	
	create-dmg \
        --volname "$(app_name)" \
        --background "./dmg_background.png" \
        --window-pos 200 120 \
        --window-size 500 332 \
        --icon-size 120 \
        --text-size 14 \
        --icon "$(app_name).app" 100 180 \
        --hide-extension "$(app_name).app" \
        --format UDBZ \
        --app-drop-link 390 180 \
        "$(app_name).dmg" \
        "./tmp"
	mv $(app_name).dmg ./release/$(app_name)-$(app_version)-amd64.dmg

windows:
	fyne-cross windows -arch amd64 -app-id $(app_id) -app-version $(app_version) -icon ./icon.png -name $(app_name)

dist:
	make darwin
	make windows
	mkdir release
	mv ./fyne-cross/bin/windows-amd64/$(app_name).exe ./release/$(app_name)-$(app_version)-amd64.exe
	mv ./fyne-cross/dist/darwin-amd64/$(app_name).app ./release/
