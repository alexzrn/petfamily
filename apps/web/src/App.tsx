import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import Layout from "./Layout";
import RequireAuth from "./RequireAuth";
import Home from "./pages/Home";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Pets from "./pages/Pets";
import MapPage from "./pages/Map";
import Walks from "./pages/Walks";
import Meetups from "./pages/Meetups";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route index element={<Home />} />
          <Route path="login" element={<Login />} />
          <Route path="register" element={<Register />} />
          <Route
            path="pets"
            element={
              <RequireAuth>
                <Pets />
              </RequireAuth>
            }
          />
          <Route path="map" element={<MapPage />} />
          <Route
            path="walks"
            element={
              <RequireAuth>
                <Walks />
              </RequireAuth>
            }
          />
          <Route path="meetups" element={<Meetups />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
