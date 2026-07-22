import { lazy, Suspense } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { VersionBadge } from '@els/ui'
import { AppShell } from './components/AppShell'
import { AuthProvider } from './auth/AuthContext'
import { RequireApp } from './auth/RequireApp'
import { RequireAuth } from './auth/RequireAuth'
import LoginPage from './pages/LoginPage'
import LoginConfirmPage from './pages/LoginConfirmPage'
import SetPasswordPage from './pages/SetPasswordPage'
import ForgotPasswordPage from './pages/ForgotPasswordPage'

const ProfileAppRoutes = lazy(() =>
  import('@els/profile-app').then((m) => ({ default: m.ProfileAppRoutes })),
)

const AdminAppRoutes = lazy(() =>
  import('@els/admin-app').then((m) => ({ default: m.AdminAppRoutes })),
)

const JournalAppRoutes = lazy(() =>
  import('@els/journal-app').then((m) => ({ default: m.JournalAppRoutes })),
)

const QuestAppRoutes = lazy(() =>
  import('@els/quest-app').then((m) => ({ default: m.QuestAppRoutes })),
)

const GrammarbookAppRoutes = lazy(() =>
  import('@els/grammarbook-app').then((m) => ({ default: m.GrammarbookAppRoutes })),
)

const EssentialbookAppRoutes = lazy(() =>
  import('@els/essentialbook-app').then((m) => ({ default: m.EssentialbookAppRoutes })),
)

const WordbookAppRoutes = lazy(() =>
  import('@els/wordbook-app').then((m) => ({ default: m.WordbookAppRoutes })),
)

const PhrasebookAppRoutes = lazy(() =>
  import('@els/phrasebook-app').then((m) => ({ default: m.PhrasebookAppRoutes })),
)

const FilmsAppRoutes = lazy(() =>
  import('@els/films-app').then((m) => ({ default: m.FilmsAppRoutes })),
)

const VocabAppRoutes = lazy(() =>
  import('@els/vocab-app').then((m) => ({ default: m.VocabAppRoutes })),
)

const ReaderAppRoutes = lazy(() =>
  import('@els/reader-app').then((m) => ({ default: m.ReaderAppRoutes })),
)

const SettingsAppRoutes = lazy(() =>
  import('@els/settings-app').then((m) => ({ default: m.SettingsAppRoutes })),
)

const ChatAppRoutes = lazy(() =>
  import('@els/chat-app').then((m) => ({ default: m.ChatAppRoutes })),
)

const SpeakingAppRoutes = lazy(() =>
  import('@els/speaking-app').then((m) => ({ default: m.SpeakingAppRoutes })),
)

const DiaryAppRoutes = lazy(() =>
  import('@els/diary-app').then((m) => ({ default: m.DiaryAppRoutes })),
)

const WritingAppRoutes = lazy(() =>
  import('@els/writing-app').then((m) => ({ default: m.WritingAppRoutes })),
)

const ReadingAppRoutes = lazy(() =>
  import('@els/reading-app').then((m) => ({ default: m.ReadingAppRoutes })),
)

const ListeningAppRoutes = lazy(() =>
  import('@els/listening-app').then((m) => ({ default: m.ListeningAppRoutes })),
)

const WorkoutAppRoutes = lazy(() =>
  import('@els/workout-app').then((m) => ({ default: m.WorkoutAppRoutes })),
)

const StudioAppRoutes = lazy(() =>
  import('@els/studio-app').then((m) => ({ default: m.StudioAppRoutes })),
)

function AppLoader() {
  return (
    <div className="flex items-center justify-center h-full w-full py-16 text-neutral-500">
      <span className="inline-block h-5 w-5 rounded-full border-2 border-neutral-300 border-t-neutral-700 animate-spin" />
    </div>
  )
}

const APP_VERSION = import.meta.env.VITE_APP_VERSION ?? 'dev'

export default function App() {
  return (
    <AuthProvider>
      <VersionBadge version={APP_VERSION} />
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/login/confirm" element={<LoginConfirmPage />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />
        <Route path="/set-password" element={<SetPasswordPage mode="set" />} />
        <Route path="/reset-password" element={<SetPasswordPage mode="reset" />} />

        <Route element={<RequireAuth />}>
          <Route element={<AppShell />}>
            <Route index element={<Navigate to="/v1/profile" replace />} />
            <Route element={<RequireApp />}>
              <Route
                path="v1/profile/*"
                element={
                  <Suspense fallback={<AppLoader />}>
                    <ProfileAppRoutes />
                  </Suspense>
                }
              />
              <Route
                path="v1/admin/*"
                element={
                  <Suspense fallback={<AppLoader />}>
                    <AdminAppRoutes />
                  </Suspense>
                }
              />
              <Route
                path="v1/journal/*"
                element={
                  <Suspense fallback={<AppLoader />}>
                    <JournalAppRoutes />
                  </Suspense>
                }
              />
              <Route
                path="v1/quest/*"
                element={
                  <Suspense fallback={<AppLoader />}>
                    <QuestAppRoutes />
                  </Suspense>
                }
              />
              <Route
                path="v1/settings/*"
                element={
                  <Suspense fallback={<AppLoader />}>
                    <SettingsAppRoutes />
                  </Suspense>
                }
              />
            </Route>
            <Route
              path="v1/grammarbook/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <GrammarbookAppRoutes />
                </Suspense>
              }
            />
            <Route
              path="v1/essentialbook/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <EssentialbookAppRoutes />
                </Suspense>
              }
            />
            <Route
              path="v1/wordbook/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <WordbookAppRoutes />
                </Suspense>
              }
            />
            <Route
              path="v1/phrasebook/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <PhrasebookAppRoutes />
                </Suspense>
              }
            />
            <Route
              path="v1/films/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <FilmsAppRoutes />
                </Suspense>
              }
            />
            <Route
              path="v1/vocab/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <VocabAppRoutes />
                </Suspense>
              }
            />
            <Route
              path="v1/reader/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <ReaderAppRoutes />
                </Suspense>
              }
            />
            <Route
              path="v1/chat/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <ChatAppRoutes />
                </Suspense>
              }
            />
            <Route
              path="v1/speaking/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <SpeakingAppRoutes />
                </Suspense>
              }
            />
            <Route
              path="v1/diary/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <DiaryAppRoutes />
                </Suspense>
              }
            />
            <Route
              path="v1/writing/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <WritingAppRoutes />
                </Suspense>
              }
            />
            <Route
              path="v1/reading/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <ReadingAppRoutes />
                </Suspense>
              }
            />
            <Route
              path="v1/listening/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <ListeningAppRoutes />
                </Suspense>
              }
            />
            <Route
              path="v1/workout/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <WorkoutAppRoutes />
                </Suspense>
              }
            />
            <Route
              path="v1/studio/*"
              element={
                <Suspense fallback={<AppLoader />}>
                  <StudioAppRoutes />
                </Suspense>
              }
            />
            <Route path="*" element={<Navigate to="/v1/profile" replace />} />
          </Route>
        </Route>
      </Routes>
    </AuthProvider>
  )
}
