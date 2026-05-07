// ==================== Flappy Bird 游戏逻辑 ====================
(function() {
    // 游戏配置
    var CANVAS_WIDTH = 400;
    var CANVAS_HEIGHT = 600;
    var GRAVITY = 0.06;
    var JUMP_STRENGTH = -2.4;
    var PIPE_SPEED = 1.2;
    var PIPE_SPAWN_RATE = 180; // 帧数
    var PIPE_GAP = 140;
    var PIPE_WIDTH = 60;
    
    // 游戏状态
    var canvas = document.getElementById('gameCanvas');
    var ctx = canvas.getContext('2d');
    var gameModal = document.getElementById('gameModal');
    var gameScore = document.getElementById('gameScore');
    var finalScore = document.getElementById('finalScore');
    var startHint = document.getElementById('gameStartHint');
    var gameOverScreen = document.getElementById('gameOverScreen');
    
    // 游戏变量
    var bird = { x: 80, y: 250, velocity: 0, radius: 16, rotation: 0, wingAngle: 0 };
    var pipes = [];
    var score = 0;
    var frameCount = 0;
    var isPlaying = false;
    var isPaused = false;
    var animationId = null;
    var gameStartTime = null;  // 游戏开始时间
    var pipesPassed = 0;       // 通过的管道数量
    
    // ==================== 音效系统 ====================
    var audioCtx = null;
    var bgmOscillators = [];
    var isMuted = false;
    
    // 初始化音频上下文（用户交互后调用）
    function initAudio() {
        if (!audioCtx) {
            audioCtx = new (window.AudioContext || window.webkitAudioContext)();
        }
        if (audioCtx.state === 'suspended') {
            audioCtx.resume();
        }
    }
    
    // 跳跃音效 - 清脆的上升音
    function playJumpSound() {
        if (!audioCtx || isMuted) return;
        try {
            var osc = audioCtx.createOscillator();
            var gain = audioCtx.createGain();
            osc.connect(gain);
            gain.connect(audioCtx.destination);
            
            // 使用方波产生清脆感
            osc.type = 'triangle';
            // 快速上升的音调
            osc.frequency.setValueAtTime(300, audioCtx.currentTime);
            osc.frequency.exponentialRampToValueAtTime(600, audioCtx.currentTime + 0.1);
            
            // 快速衰减
            gain.gain.setValueAtTime(0.3, audioCtx.currentTime);
            gain.gain.exponentialRampToValueAtTime(0.01, audioCtx.currentTime + 0.15);
            
            osc.start(audioCtx.currentTime);
            osc.stop(audioCtx.currentTime + 0.15);
        } catch(e) {}
    }
    
    // 得分音效 - 愉悦的上升旋律
    function playScoreSound() {
        if (!audioCtx || isMuted) return;
        try {
            // 创建一个和弦效果
            var frequencies = [523.25, 659.25, 783.99]; // C大调和弦 C-E-G
            frequencies.forEach(function(freq, i) {
                var osc = audioCtx.createOscillator();
                var gain = audioCtx.createGain();
                osc.connect(gain);
                gain.connect(audioCtx.destination);
                
                osc.type = 'sine';
                osc.frequency.setValueAtTime(freq, audioCtx.currentTime);
                
                // 轻微的延迟产生琶音效果
                var delay = i * 0.03;
                gain.gain.setValueAtTime(0, audioCtx.currentTime + delay);
                gain.gain.linearRampToValueAtTime(0.15, audioCtx.currentTime + delay + 0.05);
                gain.gain.exponentialRampToValueAtTime(0.01, audioCtx.currentTime + delay + 0.3);
                
                osc.start(audioCtx.currentTime + delay);
                osc.stop(audioCtx.currentTime + delay + 0.3);
            });
        } catch(e) {}
    }
    
    // 游戏结束音效 - 低沉的下降音
    function playGameOverSound() {
        if (!audioCtx || isMuted) return;
        try {
            var osc = audioCtx.createOscillator();
            var gain = audioCtx.createGain();
            osc.connect(gain);
            gain.connect(audioCtx.destination);
            
            osc.type = 'sawtooth';
            // 下降的悲伤音调
            osc.frequency.setValueAtTime(400, audioCtx.currentTime);
            osc.frequency.exponentialRampToValueAtTime(100, audioCtx.currentTime + 0.5);
            
            gain.gain.setValueAtTime(0.3, audioCtx.currentTime);
            gain.gain.exponentialRampToValueAtTime(0.01, audioCtx.currentTime + 0.5);
            
            osc.start(audioCtx.currentTime);
            osc.stop(audioCtx.currentTime + 0.5);
        } catch(e) {}
    }
    
    // 背景音乐 - 轻快的循环旋律
    function startBGM() {
        if (!audioCtx || isMuted) return;
        stopBGM(); // 先停止之前的背景音乐
        
        try {
            // 创建一个轻快的八度音程循环
            var baseFreq = 261.63; // C4
            var melody = [0, 4, 7, 4, 0, 4, 7, 12]; // 简单旋律：C-E-G-E-C-E-G-C(octave)
            var noteIndex = 0;
            
            // 使用持续的低音作为背景
            var bassOsc = audioCtx.createOscillator();
            var bassGain = audioCtx.createGain();
            bassOsc.connect(bassGain);
            bassGain.connect(audioCtx.destination);
            bassOsc.type = 'sine';
            bassOsc.frequency.value = baseFreq / 2; // 低音 C
            bassGain.gain.value = 0.05;
            bassOsc.start();
            bgmOscillators.push(bassOsc);
            
            // 创建旋律循环
            function playMelodyNote() {
                if (!isPlaying || isPaused || !audioCtx) return;
                
                var note = melody[noteIndex % melody.length];
                var freq = baseFreq * Math.pow(2, note / 12);
                
                var osc = audioCtx.createOscillator();
                var gain = audioCtx.createGain();
                osc.connect(gain);
                gain.connect(audioCtx.destination);
                
                osc.type = 'triangle';
                osc.frequency.value = freq;
                gain.gain.setValueAtTime(0.08, audioCtx.currentTime);
                gain.gain.exponentialRampToValueAtTime(0.01, audioCtx.currentTime + 0.3);
                
                osc.start(audioCtx.currentTime);
                osc.stop(audioCtx.currentTime + 0.3);
                bgmOscillators.push(osc);
                
                noteIndex++;
                // 每0.4秒一个音符
                setTimeout(playMelodyNote, 400);
            }
            
            // 延迟一点开始，避免和游戏开始音效冲突
            setTimeout(playMelodyNote, 200);
            
        } catch(e) {}
    }
    
    // 停止背景音乐
    function stopBGM() {
        bgmOscillators.forEach(function(osc) {
            try {
                osc.stop();
            } catch(e) {}
        });
        bgmOscillators = [];
    }
    
    // 静音/取消静音切换
    function toggleMute() {
        isMuted = !isMuted;
        if (isMuted) {
            stopBGM();
        } else if (isPlaying && !isPaused) {
            startBGM();
        }
        return isMuted;
    }
    
    // 配色方案（根据主题自动适应）
    function getColors() {
        var isLight = document.documentElement.classList.contains('light');
        if (isLight) {
            return {
                bg: ['#87CEEB', '#E0F6FF'],
                pipe: '#2d5a27',
                pipeCap: '#3d7a37',
                bird: '#e8a86d',
                birdEye: '#fff',
                birdBeak: '#f5a623',
                ground: '#ded895',
                groundTop: '#d4c876'
            };
        }
        return {
            bg: ['#1a2332', '#0d1218'],
            pipe: '#2d3436',
            pipeCap: '#636e72',
            bird: '#e8a86d',
            birdEye: '#fff',
            birdBeak: '#f5a623',
            ground: '#2d3436',
            groundTop: '#636e72'
        };
    }
    
    // 绘制背景
    function drawBackground() {
        var colors = getColors();
        var gradient = ctx.createLinearGradient(0, 0, 0, CANVAS_HEIGHT);
        gradient.addColorStop(0, colors.bg[0]);
        gradient.addColorStop(1, colors.bg[1]);
        ctx.fillStyle = gradient;
        ctx.fillRect(0, 0, CANVAS_WIDTH, CANVAS_HEIGHT);
        
        // 绘制星星/云朵（装饰）
        ctx.fillStyle = 'rgba(255,255,255,0.15)';
        for (var i = 0; i < 8; i++) {
            var x = (frameCount * 0.3 + i * 50) % (CANVAS_WIDTH + 50) - 25;
            var y = 50 + i * 35;
            ctx.beginPath();
            ctx.arc(x, y, 3, 0, Math.PI * 2);
            ctx.fill();
        }
    }
    
    // 绘制地面
    function drawGround() {
        var colors = getColors();
        var groundHeight = 50;
        
        // 地面主体
        ctx.fillStyle = colors.ground;
        ctx.fillRect(0, CANVAS_HEIGHT - groundHeight, CANVAS_WIDTH, groundHeight);
        
        // 地面顶部线条
        ctx.fillStyle = colors.groundTop;
        ctx.fillRect(0, CANVAS_HEIGHT - groundHeight, CANVAS_WIDTH, 4);
        
        // 地面纹理
        ctx.fillStyle = 'rgba(0,0,0,0.1)';
        for (var i = 0; i < CANVAS_WIDTH; i += 20) {
            var offset = (frameCount * PIPE_SPEED) % 20;
            ctx.fillRect(i - offset, CANVAS_HEIGHT - groundHeight + 10, 10, 3);
            ctx.fillRect(i - offset + 5, CANVAS_HEIGHT - groundHeight + 25, 8, 2);
        }
    }
    
    // 绘制小鸟
    function drawBird() {
        var colors = getColors();
        
        ctx.save();
        ctx.translate(bird.x, bird.y);
        ctx.rotate(bird.rotation);
        
        // 身体
        ctx.fillStyle = colors.bird;
        ctx.beginPath();
        ctx.arc(0, 0, bird.radius, 0, Math.PI * 2);
        ctx.fill();
        
        // 身体高光
        ctx.fillStyle = 'rgba(255,255,255,0.2)';
        ctx.beginPath();
        ctx.arc(-4, -4, 6, 0, Math.PI * 2);
        ctx.fill();
        
        // 眼睛
        ctx.fillStyle = colors.birdEye;
        ctx.beginPath();
        ctx.arc(6, -4, 5, 0, Math.PI * 2);
        ctx.fill();
        
        // 眼珠
        ctx.fillStyle = '#333';
        ctx.beginPath();
        ctx.arc(8, -4, 2.5, 0, Math.PI * 2);
        ctx.fill();
        
        // 眼珠高光
        ctx.fillStyle = '#fff';
        ctx.beginPath();
        ctx.arc(9, -5, 1, 0, Math.PI * 2);
        ctx.fill();
        
        // 嘴巴
        ctx.fillStyle = colors.birdBeak;
        ctx.beginPath();
        ctx.moveTo(10, 2);
        ctx.lineTo(18, 5);
        ctx.lineTo(10, 8);
        ctx.closePath();
        ctx.fill();
        
        // 翅膀 - 动态扇动效果
        ctx.save();
        ctx.translate(-6, 2);
        ctx.rotate(bird.wingAngle);
        ctx.fillStyle = '#d4956d';
        ctx.beginPath();
        ctx.ellipse(0, 0, 9, 6, 0, 0, Math.PI * 2);
        ctx.fill();
        // 翅膀阴影
        ctx.fillStyle = 'rgba(0,0,0,0.1)';
        ctx.beginPath();
        ctx.ellipse(2, 2, 9, 6, 0, 0, Math.PI * 2);
        ctx.fill();
        ctx.restore();
        
        ctx.restore();
    }
    
    // 绘制管道
    function drawPipe(pipe) {
        var colors = getColors();
        
        // 管道渐变
        var gradient = ctx.createLinearGradient(pipe.x, 0, pipe.x + PIPE_WIDTH, 0);
        gradient.addColorStop(0, colors.pipe);
        gradient.addColorStop(0.5, colors.pipeCap);
        gradient.addColorStop(1, colors.pipe);
        
        // 上管道
        ctx.fillStyle = gradient;
        ctx.fillRect(pipe.x, 0, PIPE_WIDTH, pipe.topHeight);
        
        // 上管道边缘
        ctx.fillStyle = colors.pipeCap;
        ctx.fillRect(pipe.x - 3, pipe.topHeight - 20, PIPE_WIDTH + 6, 20);
        
        // 下管道
        ctx.fillStyle = gradient;
        ctx.fillRect(pipe.x, pipe.topHeight + PIPE_GAP, PIPE_WIDTH, CANVAS_HEIGHT - pipe.topHeight - PIPE_GAP - 50);
        
        // 下管道边缘
        ctx.fillStyle = colors.pipeCap;
        ctx.fillRect(pipe.x - 3, pipe.topHeight + PIPE_GAP, PIPE_WIDTH + 6, 20);
        
        // 管道高光
        ctx.fillStyle = 'rgba(255,255,255,0.1)';
        ctx.fillRect(pipe.x + 5, 0, 4, pipe.topHeight);
        ctx.fillRect(pipe.x + 5, pipe.topHeight + PIPE_GAP, 4, CANVAS_HEIGHT - pipe.topHeight - PIPE_GAP - 50);
    }
    
    // 更新游戏
    function update() {
        if (!isPlaying || isPaused) return;
        
        frameCount++;
        
        // 更新小鸟
        bird.velocity += GRAVITY;
        bird.y += bird.velocity;
        
        // 更新旋转角度
        bird.rotation = Math.min(Math.PI / 4, Math.max(-Math.PI / 4, bird.velocity * 0.1));
        
        // 翅膀扇动动画 - 上升时快速扇动，下落时缓慢
        if (bird.velocity < 0) {
            // 上升时快速扇动
            bird.wingAngle = Math.sin(frameCount * 0.4) * 0.6;
        } else {
            // 下落时缓慢扇动
            bird.wingAngle = Math.sin(frameCount * 0.1) * 0.3;
        }
        
        // 检查地面碰撞
        if (bird.y + bird.radius >= CANVAS_HEIGHT - 50) {
            bird.y = CANVAS_HEIGHT - 50 - bird.radius;
            gameOver();
            return;
        }
        
        // 检查天花板碰撞
        if (bird.y - bird.radius <= 0) {
            bird.y = bird.radius;
            bird.velocity = 0;
        }
        
        // 生成管道
        if (frameCount % PIPE_SPAWN_RATE === 0) {
            var minHeight = 50;
            var maxHeight = CANVAS_HEIGHT - PIPE_GAP - minHeight - 50;
            var topHeight = Math.floor(Math.random() * (maxHeight - minHeight) + minHeight);
            pipes.push({
                x: CANVAS_WIDTH,
                topHeight: topHeight,
                passed: false
            });
        }
        
        // 更新管道
        for (var i = pipes.length - 1; i >= 0; i--) {
            var pipe = pipes[i];
            pipe.x -= PIPE_SPEED;
            
            // 计分
            if (!pipe.passed && pipe.x + PIPE_WIDTH < bird.x) {
                pipe.passed = true;
                score++;
                gameScore.textContent = score;
                playScoreSound(); // 播放得分音效
            }
            
            // 碰撞检测
            if (bird.x + bird.radius > pipe.x && bird.x - bird.radius < pipe.x + PIPE_WIDTH) {
                if (bird.y - bird.radius < pipe.topHeight || bird.y + bird.radius > pipe.topHeight + PIPE_GAP) {
                    gameOver();
                    return;
                }
            }
            
            // 移除屏幕外管道
            if (pipe.x + PIPE_WIDTH < 0) {
                pipes.splice(i, 1);
            }
        }
    }
    
    // 绘制游戏
    function draw() {
        ctx.clearRect(0, 0, CANVAS_WIDTH, CANVAS_HEIGHT);
        
        drawBackground();
        
        // 绘制管道
        pipes.forEach(drawPipe);
        
        drawGround();
        drawBird();
    }
    
    // 游戏循环
    function gameLoop() {
        update();
        draw();
        animationId = requestAnimationFrame(gameLoop);
    }
    
    // 跳跃
    function jump() {
        if (!isPlaying) {
            startGame();
            return;
        }
        if (isPaused) return;
        bird.velocity = JUMP_STRENGTH;
        playJumpSound(); // 播放跳跃音效
    }
    
    // 开始游戏
    function startGame() {
        // 初始化音频系统（需要用户交互后才能播放）
        initAudio();
        
        // 先取消之前的动画帧，防止多个循环同时运行
        cancelAnimationFrame(animationId);
        isPlaying = true;
        isPaused = false;
        gameStartTime = Date.now();  // 记录游戏开始时间
        startHint.classList.add('hide');
        gameOverScreen.classList.remove('show');
        // 恢复画布交互
        canvas.style.pointerEvents = 'auto';
        resetGame();
        gameLoop();
        
        // 开始背景音乐
        startBGM();
    }
    
    // 重置游戏
    function resetGame() {
        bird.y = 250;
        bird.velocity = 0;
        bird.rotation = 0;
        bird.wingAngle = 0;
        pipes = [];
        score = 0;
        pipesPassed = 0;
        frameCount = 0;
        gameScore.textContent = '0';
    }
    
    // 格式化游戏时长为 mm:ss
    function formatGameTime(ms) {
        var seconds = Math.floor(ms / 1000);
        var mins = Math.floor(seconds / 60);
        var secs = seconds % 60;
        return (mins < 10 ? '0' : '') + mins + ':' + (secs < 10 ? '0' : '') + secs;
    }
    
    // 历史最高分
    var highScore = 0;
    
    // 从本地存储加载最高分
    function loadHighScore() {
        var saved = localStorage.getItem('flappyHighScore');
        if (saved) {
            highScore = parseInt(saved, 10);
        }
    }
    
    // 保存最高分
    function saveHighScore() {
        if (score > highScore) {
            highScore = score;
            localStorage.setItem('flappyHighScore', highScore);
        }
    }
    
    // 游戏结束
    function gameOver() {
        isPlaying = false;
        cancelAnimationFrame(animationId);
        
        // 停止背景音乐，播放游戏结束音效
        stopBGM();
        playGameOverSound();
        
        // 更新并保存最高分
        saveHighScore();
        
        // 显示结果展示页
        showResultScreen();
    }
    
    // 显示结果展示页
    function showResultScreen() {
        // 计算游戏时长
        var gameDuration = gameStartTime ? Date.now() - gameStartTime : 0;
        
        finalScore.textContent = score;
        
        // 更新最高分显示
        var highScoreEl = document.getElementById('highScore');
        var newRecordEl = document.getElementById('newRecord');
        var pipesPassedEl = document.getElementById('pipesPassed');
        var gameTimeEl = document.getElementById('gameTime');
        
        if (highScoreEl) {
            highScoreEl.textContent = highScore;
        }
        
        // 显示通过的管道数（得分就是通过的管道数）
        if (pipesPassedEl) {
            pipesPassedEl.textContent = score;
        }
        
        // 显示游戏时长
        if (gameTimeEl) {
            gameTimeEl.textContent = formatGameTime(gameDuration);
        }
        
        // 显示"新纪录"标记（如果破纪录了）
        if (newRecordEl) {
            if (score > 0 && score > highScore) {
                newRecordEl.style.display = 'inline-block';
            } else {
                newRecordEl.style.display = 'none';
            }
        }
        
        // 禁用游戏画布交互，防止空格键误触
        canvas.style.pointerEvents = 'none';
        
        gameOverScreen.classList.add('show');
    }
    
    // 重新开始
    function restart() {
        gameOverScreen.classList.remove('show');
        // 恢复画布交互
        canvas.style.pointerEvents = 'auto';
        startGame();
    }
    
    // 暂停/继续
    function togglePause() {
        if (!isPlaying) return;
        isPaused = !isPaused;
    }
    
    // 打开游戏弹窗
    function openGame() {
        gameModal.classList.add('show');
        resetGame();
        draw();
    }
    
    // 关闭游戏弹窗
    function closeGame() {
        gameModal.classList.remove('show');
        isPlaying = false;
        isPaused = false;
        cancelAnimationFrame(animationId);
        stopBGM(); // 停止背景音乐
        startHint.classList.remove('hide');
        gameOverScreen.classList.remove('show');
    }
    
    // 事件绑定
    document.getElementById('game-trigger').addEventListener('click', function(e) {
        e.preventDefault();
        openGame();
    });
    
    document.getElementById('gameCloseBtn').addEventListener('click', closeGame);
    document.getElementById('gameRestartBtn').addEventListener('click', restart);
    
    // 关闭结果页按钮 - 只关闭结果展示，回到准备状态
    var closeResultBtn = document.getElementById('gameCloseResultBtn');
    if (closeResultBtn) {
        closeResultBtn.addEventListener('click', function() {
            gameOverScreen.classList.remove('show');
            // 恢复画布交互
            canvas.style.pointerEvents = 'auto';
            // 重置到准备状态（显示开始提示）
            isPlaying = false;
            isPaused = false;
            resetGame();
            draw();
            startHint.classList.remove('hide');
        });
    }
    
    // 画布点击/触摸 - 使用mousedown响应更快
    canvas.addEventListener('mousedown', function(e) {
        e.preventDefault();
        jump();
    });
    
    canvas.addEventListener('touchstart', function(e) {
        e.preventDefault();
        jump();
    }, { passive: false });
    
    // 键盘控制
    document.addEventListener('keydown', function(e) {
        if (!gameModal.classList.contains('show')) return;
        
        if (e.code === 'Space') {
            e.preventDefault();
            jump();
        } else if (e.code === 'KeyP') {
            e.preventDefault();
            togglePause();
        } else if (e.code === 'Escape') {
            closeGame();
        }
    });
    
    // 点击画布外部关闭
    gameModal.addEventListener('click', function(e) {
        if (e.target === gameModal) {
            closeGame();
        }
    });
    
    // 初始绘制
    draw();
})();
