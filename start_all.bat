@echo off
echo ================================================
echo   GREEN Network Energy Monitor - Startup
echo ================================================

echo Starting Program A (Power and Energy) on port 8881...
start "Program A" cmd /k "cd /d "C:\QAMROSH MAQSOOD\Local Disk E\Master in IT (IOT)\sonoff\green_models\program_a" && go run ."

echo Starting Program B (SmartPDU) on port 8882...
start "Program B" cmd /k "cd /d "C:\QAMROSH MAQSOOD\Local Disk E\Master in IT (IOT)\sonoff\green_models\program_b" && go run ."

echo Starting Program C (ISAC Utilization) on port 8883...
start "Program C" cmd /k "cd /d "C:\QAMROSH MAQSOOD\Local Disk E\Master in IT (IOT)\sonoff\green_models\program_c" && go run ."

echo Starting Sonoff Server on port 8080...
start "Sonoff" cmd /k "cd /d "C:\QAMROSH MAQSOOD\Local Disk E\Master in IT (IOT)\sonoff\restconf\cmd\rc-test-server" && go run ."

timeout /t 3 /nobreak

echo Starting Aggregator on port 8884...
start "Aggregator" cmd /k "cd /d "C:\QAMROSH MAQSOOD\Local Disk E\Master in IT (IOT)\sonoff\green_models\aggregator" && go run ."

timeout /t 3 /nobreak

echo Starting Data Scheduler...
start "Scheduler" cmd /k "cd /d "C:\QAMROSH MAQSOOD\Local Disk E\Master in IT (IOT)\sonoff\green_models" && python scheduler.py"

echo ================================================
echo   All servers started!
echo   Open dashboard.html in your browser
echo ================================================
pause
