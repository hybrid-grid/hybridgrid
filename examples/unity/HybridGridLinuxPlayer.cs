using System;
using System.IO;
using System.Linq;
using UnityEditor;
using UnityEditor.Build;
using UnityEditor.Build.Reporting;

public static class HybridGridLinuxPlayer
{
    public static void Build()
    {
        var scenes = EditorBuildSettings.scenes.Where(scene => scene.enabled)
                                              .Select(scene => scene.path).ToArray();
        if (scenes.Length == 0)
            throw new InvalidOperationException("Enable at least one scene in Build Settings.");

        var target = NamedBuildTarget.Standalone;
        var previousBackend = PlayerSettings.GetScriptingBackend(target);
        try
        {
            PlayerSettings.SetScriptingBackend(target, ScriptingImplementation.IL2CPP);
            Directory.CreateDirectory("Build");
            var report = BuildPipeline.BuildPlayer(new BuildPlayerOptions
            {
                scenes = scenes,
                locationPathName = "Build/Game.x86_64",
                target = BuildTarget.StandaloneLinux64,
                options = BuildOptions.None
            });
            if (report.summary.result != BuildResult.Succeeded)
                throw new InvalidOperationException("Linux IL2CPP Player build " + report.summary.result);
        }
        finally
        {
            PlayerSettings.SetScriptingBackend(target, previousBackend);
        }
    }
}
