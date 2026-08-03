// Эффект слежения за курсором с сердечками
class CursorEffect {
    constructor() {
        this.particles = [];
        this.mouseX = 0;
        this.mouseY = 0;
        this.isMouseOver = false;
        this.canvas = null;
        this.ctx = null;
        this.animationId = null;
        this.lastParticleTime = 0;
        this.particleDelay = 80; // задержка между частицами (мс)
        this.init();
    }

    init() {
        // Создаём canvas поверх всего
        this.canvas = document.createElement('canvas');
        this.canvas.style.position = 'fixed';
        this.canvas.style.top = '0';
        this.canvas.style.left = '0';
        this.canvas.style.pointerEvents = 'none';
        this.canvas.style.zIndex = '9999';
        this.canvas.width = window.innerWidth;
        this.canvas.height = window.innerHeight;
        document.body.appendChild(this.canvas);

        this.ctx = this.canvas.getContext('2d');

        // Слушаем движение мыши
        document.addEventListener('mousemove', (e) => {
            this.mouseX = e.clientX;
            this.mouseY = e.clientY;
            this.isMouseOver = true;
            
            // Создаём частицы с задержкой
            const now = Date.now();
            if (now - this.lastParticleTime > this.particleDelay) {
                this.createParticles(1); // только 1 сердечко за раз
                this.lastParticleTime = now;
            }
        });

        document.addEventListener('mouseleave', () => {
            this.isMouseOver = false;
        });

        window.addEventListener('resize', () => {
            this.canvas.width = window.innerWidth;
            this.canvas.height = window.innerHeight;
        });

        this.animate();
    }

    createParticles(count) {
        const colors = [
            '#FF6B6B', '#FF4757', '#FF6348', '#FF9F43',
            '#F368E0', '#FF9FF3', '#FF6B81', '#FF4757',
            '#EE5A24', '#FF3838'
        ];

        for (let i = 0; i < count; i++) {
            const size = Math.random() * 20 + 18; // 18-38px
            this.particles.push({
                x: this.mouseX + (Math.random() - 0.5) * 20,
                y: this.mouseY + (Math.random() - 0.5) * 20,
                size: size,
                speedX: (Math.random() - 0.5) * 4,
                speedY: (Math.random() - 0.5) * 4 - 1.5,
                life: 1,
                decay: Math.random() * 0.012 + 0.008, // живут дольше
                color: colors[Math.floor(Math.random() * colors.length)],
                rotation: Math.random() * Math.PI * 2,
                rotationSpeed: (Math.random() - 0.5) * 0.08,
                scale: Math.random() * 0.3 + 0.8
            });
        }
    }

    drawHeart(ctx, x, y, size, rotation, color, alpha) {
        ctx.save();
        ctx.translate(x, y);
        ctx.rotate(rotation);
        ctx.globalAlpha = alpha;
        ctx.shadowColor = 'rgba(255, 71, 87, 0.5)';
        ctx.shadowBlur = 20;

        // Рисуем сердечко
        ctx.beginPath();
        ctx.moveTo(0, size * 0.3);
        ctx.bezierCurveTo(-size * 0.6, -size * 0.3, -size * 0.8, size * 0.2, 0, size * 0.8);
        ctx.bezierCurveTo(size * 0.8, size * 0.2, size * 0.6, -size * 0.3, 0, size * 0.3);
        ctx.closePath();
        
        // Градиент для сердечка
        const gradient = ctx.createRadialGradient(
            -size * 0.2, -size * 0.2, 0,
            -size * 0.2, -size * 0.2, size
        );
        gradient.addColorStop(0, '#FFD93D');
        gradient.addColorStop(0.3, color);
        gradient.addColorStop(1, '#FF6B6B');
        
        ctx.fillStyle = gradient;
        ctx.fill();
        
        // Белый блик
        ctx.shadowBlur = 0;
        ctx.beginPath();
        ctx.arc(-size * 0.2, -size * 0.3, size * 0.15, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(255, 255, 255, ${alpha * 0.5})`;
        ctx.fill();

        ctx.restore();
    }

    animate() {
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

        // Обновляем и рисуем частицы
        for (let i = this.particles.length - 1; i >= 0; i--) {
            const p = this.particles[i];
            
            // Обновляем позицию
            p.x += p.speedX;
            p.y += p.speedY;
            p.speedX *= 0.98;
            p.speedY *= 0.98;
            p.speedY += 0.03; // лёгкое падение
            p.life -= p.decay;
            p.rotation += p.rotationSpeed;

            // Рисуем частицу
            if (p.life > 0) {
                const alpha = p.life * 0.9;
                this.drawHeart(
                    this.ctx,
                    p.x,
                    p.y,
                    p.size * p.life * p.scale,
                    p.rotation,
                    p.color,
                    alpha
                );
            } else {
                this.particles.splice(i, 1);
            }
        }

        // Рисуем лёгкий след за курсором
        if (this.isMouseOver) {
            const gradient = this.ctx.createRadialGradient(
                this.mouseX, this.mouseY, 0,
                this.mouseX, this.mouseY, 40
            );
            gradient.addColorStop(0, 'rgba(255, 107, 107, 0.15)');
            gradient.addColorStop(0.5, 'rgba(255, 71, 87, 0.05)');
            gradient.addColorStop(1, 'rgba(255, 71, 87, 0)');
            
            this.ctx.beginPath();
            this.ctx.arc(this.mouseX, this.mouseY, 40, 0, Math.PI * 2);
            this.ctx.fillStyle = gradient;
            this.ctx.fill();
        }

        // Ограничиваем количество частиц (максимум 50)
        if (this.particles.length > 50) {
            this.particles.splice(0, this.particles.length - 50);
        }

        this.animationId = requestAnimationFrame(() => this.animate());
    }
}

// Инициализируем, когда страница загружена
document.addEventListener('DOMContentLoaded', () => {
    new CursorEffect();
});