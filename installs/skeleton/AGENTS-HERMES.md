# Hermes + NOUS — Enhanced Engineering

## Skills
Carga el bundle `/nous` para activar todas las skills de NOUS (13 skills).
Lista completa en `~/.hermes/OKF/_system/conventions.md`.

## Workflow estándar

```
1. /grill-with-docs    → alinear requisitos + construir CONTEXT.md + ADRs
2. /openspec            → crear proposal.md + tasks.md en openspec/changes/
3. /writing-plans       → expandir a plan bite-sized (2-5 min por tarea)
4. subagent-driven-dev  → ejecutar con TDD + review 2-stage
5. /code-review         → verificar diff antes de push
```

O todo el flujo de una vez con `/nous-pipeline` (orquesta grill → openspec → plan → TDD → review con `delegate_task`).

Para trabajo enorme que no cabe en una sesión: `/wayfinder` → mapa de decision tickets.

## Al iniciar sesión

1. Leer `.agents/MEMORY.md` (contexto activo y blockers)
2. Leer `~/.hermes/OKF/index.md` (catálogo global) → seguir solo links relevantes a la tarea
3. Usar `memory` tool para hechos frecuentes (se inyecta automático en cada turno)
4. Usar `session_search("query")` para recuperar contexto de conversaciones pasadas

Los protocolos globales (idioma, OKF, git safety, backups) viven en `~/.hermes/SOUL.md` — se inyectan en cada turno automáticamente.

## Memoria y OKF

- **Activa**: Hermes `memory` tool — persiste hechos entre sesiones, inyectados cada turno
- **Durable**: `~/.hermes/OKF/<project>/` — arquitectura, decisiones, runbooks, troubleshooting
- **Local**: `.agents/OKF/` — referencia para otros agentes, gestionado por `nous sync`
- **Mantenimiento**: después de descubrimientos significativos, decisiones, comandos verificados o problemas resueltos, actualizar el concepto OKF correspondiente. Registrar milestones en `OKF/log.md` (YYYY-MM-DD, newest first). No duplicar conocimiento durable en MEMORY.md.

## Herramientas Hermes clave

| Herramienta | Cuándo usarla |
|---|---|
| `delegate_task` | Subagentes paralelos para investigación, code review, testing pesado |
| `session_search` | Buscar conversaciones pasadas con FTS5 |
| `memory` | Persistir preferencias, lecciones, hechos entre sesiones |
| `cronjob` | Tareas programadas recurrentes (watchers, reportes, monitoreo) |
| `clarify` | Preguntar al usuario cuando hay ambigüedad o decisiones con trade-offs |
| `process` | Gestionar procesos en background (servidores, builds largos) |

## Protocolos de seguridad

- **Git**: no commit ni push sin "YES" explícito del usuario tras mostrar `git diff`
- **Backups**: antes de editar cualquier archivo fuera de `dev/sandbox/`, crear copia en `dev/backups/YYYYMMDD_HHMMSS_filename.ext`
- **Rollback**: si se detectan fallos post-edición, analizar diferencias con el backup y proponer rollback con diff. No ejecutar sin confirmación.
- **Credenciales**: solo en `.env`, nunca hardcodeadas. Cargar vía variables de entorno.
- **Husky hooks**: usar `--no-verify` o `HUSKY=0` si los hooks bloquean operaciones legítimas.

## Convenciones

- **tmp-repos**: `~/Downloads/tmp-repos/` — canonical folder para clonar repos a analizar
- **Idioma**: español en conversación, inglés en código/commits/docs/tests/UI strings
- **Dependencias Python**: `uv` exclusivamente, nunca `pip`. Activar `.venv` antes de comandos
- **Estructura**: `dev/` y `.agents/` NO se trackean. `docs/` (ADRs) y `src/` SÍ.

## Estructura del proyecto

```
proyecto/
├── AGENTS.md                ← este archivo
├── dev/                     ← NO TRACKED: sandbox, tmp-repos, docs, scripts, tests, backups
├── .agents/                 ← NO TRACKED: MEMORY.md, OKF/, skills/
├── openspec/                ← specs + changes (spec-driven development)
├── docs/                    ← TRACKED: ADRs
└── src/                     ← TRACKED: código
```
